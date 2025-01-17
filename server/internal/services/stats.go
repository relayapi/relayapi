package services

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/term"

	"relayapi/server/internal/config"
	"relayapi/server/internal/middleware/logger"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type Stats struct {
	TotalRequests      uint64
	SuccessfulRequests uint64
	FailedRequests     uint64
	BytesReceived      uint64
	BytesSent          uint64
	StartTime          time.Time
	errorStats         sync.Map                       // 用于存储每个错误状态码的计数
	Version            string                         // 版本号
	ServerAddr         string                         // 服务器地址
	Clients            map[string]config.ClientConfig // 客户端配置
}

func NewStats(version, serverAddr string, clients map[string]config.ClientConfig) *Stats {
	return &Stats{
		StartTime:  time.Now(),
		Version:    version,
		ServerAddr: serverAddr,
		Clients:    clients,
	}
}

// GetUptime 返回服务器运行时间
func (s *Stats) GetUptime() time.Duration {
	return time.Since(s.StartTime)
}

func (s *Stats) IncrementTotal() {
	atomic.AddUint64(&s.TotalRequests, 1)
}

func (s *Stats) IncrementSuccess() {
	atomic.AddUint64(&s.SuccessfulRequests, 1)
}

func (s *Stats) IncrementFailed() {
	atomic.AddUint64(&s.FailedRequests, 1)
}

// IncrementErrorStatus 增加特定错误状态码的计数
func (s *Stats) IncrementErrorStatus(statusCode int) {
	if value, ok := s.errorStats.Load(statusCode); ok {
		atomic.AddUint64(value.(*uint64), 1)
	} else {
		var counter uint64 = 1
		s.errorStats.Store(statusCode, &counter)
	}
}

// GetErrorStats 获取错误状态码统计
func (s *Stats) GetErrorStats() map[int]uint64 {
	stats := make(map[int]uint64)
	s.errorStats.Range(func(key, value interface{}) bool {
		stats[key.(int)] = atomic.LoadUint64(value.(*uint64))
		return true
	})
	return stats
}

func (s *Stats) AddBytesReceived(n uint64) {
	atomic.AddUint64(&s.BytesReceived, n)
}

func (s *Stats) AddBytesSent(n uint64) {
	atomic.AddUint64(&s.BytesSent, n)
}

// 获取错误状态码的描述
func getStatusCodeDesc(code int) string {
	switch code {
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 405:
		return "Method Not Allowed"
	case 408:
		return "Request Timeout"
	case 429:
		return "Too Many Requests"
	case 500:
		return "Internal Server Error"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	case 504:
		return "Gateway Timeout"
	default:
		return "Unknown Error"
	}
}

// 格式化字节大小
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// StartConsoleDisplay 开始在控制台显示实时统计信息
func (s *Stats) StartConsoleDisplay(stopChan chan struct{}) {
	// 清屏并将光标移到开头
	fmt.Print("\033[2J\033[H")

	// 渐变色数组
	gradientColors := []string{
		"\033[38;5;51m", // 浅青色
		"\033[38;5;45m", // 青色
		"\033[38;5;39m", // 深青色
		"\033[38;5;33m", // 蓝色
		"\033[38;5;27m", // 深蓝色
	}

	// 先显示 Logo
	logo := `
    ██████╗ ███████╗██╗      █████╗ ██╗   ██╗
    ██╔══██╗██╔════╝██║     ██╔══██╗╚██╗ ██╔╝
    ██████╔╝█████╗  ██║     ███████║ ╚████╔╝ 
    ██╔══██╗██╔══╝  ██║     ██╔══██║  ╚██╔╝  
    ██║  ██║███████╗███████╗██║  ██║   ██║   
    ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝  ╚═╝   ╚═╝   
     █████╗ ██████╗ ██╗
    ██╔══██╗██╔══██╗██║
    ███████║██████╔╝██║
    ██╔══██║██╔═══╝ ██║
    ██║  ██║██║     ██║
    ╚═╝  ╚═╝╚═╝     ╚═╝ v` + s.Version + "\n"

	// 使用渐变色一次性显示 Logo
	logoLines := strings.Split(logo, "\n")
	for _, line := range logoLines {
		if len(strings.TrimSpace(line)) > 0 {
			// 为每一行应用双重渐变色
			chars := []rune(line)
			midPoint := len(chars) / 2

			// 前半部分使用一组渐变色
			for i := 0; i < midPoint; i++ {
				color := gradientColors[i%(len(gradientColors))]
				fmt.Print(color, string(chars[i]))
			}

			// 后半部分使用反向渐变色
			for i := midPoint; i < len(chars); i++ {
				color := gradientColors[(len(chars)-i)%(len(gradientColors))]
				fmt.Print(color, string(chars[i]))
			}
		}
		fmt.Print("\033[0m\n")
	}

	// 添加一个分隔线
	fmt.Print("\n")
	divider := "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	for _, char := range divider {
		colorIdx := 0
		fmt.Print(gradientColors[colorIdx%len(gradientColors)], string(char))
		colorIdx++
	}
	fmt.Print("\033[0m\n")

	// 等待一小段时间让用户欣赏 Logo
	time.Sleep(300 * time.Millisecond)

	// 显示启动信息
	fmt.Println("\n=== RelayAPI 服务启动中 ===")

	// 使用打字机效果显示服务器信息
	serverInfo := fmt.Sprintf("🚀 启动地址: %s", s.ServerAddr)
	for _, char := range serverInfo {
		fmt.Print("\033[33m", string(char), "\033[0m")
		time.Sleep(20 * time.Millisecond)
	}
	fmt.Println()

	// 使用动画效果显示初始化提示
	initText := "系统核心初始化..."
	fmt.Print("\n")
	for i := 0; i < 3; i++ {
		for _, char := range []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"} {
			fmt.Printf("\r\033[32m%s %s\033[0m", char, initText)
			time.Sleep(50 * time.Millisecond)
		}
	}
	fmt.Println("\n")

	// 显示进度条
	width := 40
	spinChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinIdx := 0

	for i := 0; i <= width; i++ {
		progress := float64(i) / float64(width) * 100
		filled := repeat('▓', i)
		empty := repeat('░', width-i)

		// 使用渐变色进度条
		colorIdx := int(float64(i) / float64(width) * float64(len(gradientColors)))
		if colorIdx >= len(gradientColors) {
			colorIdx = len(gradientColors) - 1
		}

		// 使用彩色输出和加载动画
		fmt.Printf("\r%s %s[%s%s]\033[0m \033[33m%.1f%%\033[0m",
			spinChars[spinIdx],
			gradientColors[colorIdx],
			string(filled),
			string(empty),
			progress)

		spinIdx = (spinIdx + 1) % len(spinChars)
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Println("\n")

	// 显示启动检查项，使用动画效果
	checkItems := []struct {
		text    string
		color   string
		symbols []string
	}{
		{"日志系统加载完毕", "\033[32m", []string{"⋯", "⋱", "⋮", "⋰"}},   // 绿色
		{"代理服务初始化成功", "\033[36m", []string{"◢", "◣", "◤", "◥"}},  // 青色
		{"API 处理模块就绪", "\033[33m", []string{"◐", "◓", "◑", "◒"}}, // 黄色
		{"配置流量限制中...", "\033[35m", []string{"▖", "▘", "▝", "▗"}}, // 紫色
		{"流量限制规则已部署", "\033[34m", []string{"⠋", "⠙", "⠸", "⠴"}},  // 蓝色
		{"限流中间件启动完成", "\033[32m", []string{"⣾", "⣽", "⣻", "⢿"}},  // 绿色
		{"安全认证模块已加载", "\033[36m", []string{"◢", "◣", "◤", "◥"}},  // 青色
	}

	for _, item := range checkItems {
		// 显示加载动画
		for j := 0; j < 6; j++ {
			fmt.Printf("\r%s%s %s\033[0m",
				item.color,
				item.symbols[j%len(item.symbols)],
				item.text)
			time.Sleep(50 * time.Millisecond)
		}
		// 显示完成标记
		fmt.Printf("\r%s✓ %s\033[0m\n", item.color, item.text)
		time.Sleep(100 * time.Millisecond)
	}

	// 显示服务器启动信息
	fmt.Printf("\n%s🚀 服务启动端口 %s:8840%s\n\n",
		"\033[36m", s.ServerAddr, "\033[0m")

	// 启动提示使用渐变动画
	startText := "正在初始化统计界面..."
	spinIdx = 0
	for i := 0; i < 8; i++ { // 缩短动画时间
		color := gradientColors[i%len(gradientColors)]
		fmt.Printf("\r%s%s %s\033[0m", color, spinChars[spinIdx], startText)
		spinIdx = (spinIdx + 1) % len(spinChars)
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Print("\n\n")

	var uiActive bool = true
	var uiQuit bool = false

	// 创建一个函数来启动 UI
	startUI := func() error {
		if err := ui.Init(); err != nil {
			log.Printf("failed to initialize termui: %v", err)
			return err
		}
		uiActive = true
		return nil
	}

	// 初始启动 UI
	if err := startUI(); err != nil {
		fmt.Println("启动统计界面失败，将以普通模式运行")
		return
	}
	defer ui.Close()

	// 创建标题
	title := widgets.NewParagraph()
	title.Title = fmt.Sprintf("RelayAPI Server (v%s | %s)", s.Version, s.ServerAddr)

	// 添加客户端到列表
	var clientKeys []string
	clientDetails := make(map[string]string)
	var titleText strings.Builder
	for hash, client := range s.Clients {
		shortHash := hash[:12] + "..."
		clientKeys = append(clientKeys, shortHash)
		// 存储详细信息
		maskedKey := client.Crypto.AESKey[:8] + "..." + client.Crypto.AESKey[len(client.Crypto.AESKey)-4:]
		titleText.WriteString(fmt.Sprintf("%s | Key: %s | IV: %s\n", shortHash, maskedKey, client.Crypto.AESIVSeed))
		clientDetails[shortHash] = fmt.Sprintf("Hash: %s\nKey: %s\nIV: %s", hash, maskedKey, client.Crypto.AESIVSeed)
	}
	title.Text = titleText.String()
	title.TextStyle.Fg = ui.ColorCyan
	title.BorderStyle.Fg = ui.ColorCyan
	title.TitleStyle.Fg = ui.ColorCyan

	// 创建详细信息显示区域
	clientDetail := widgets.NewParagraph()
	clientDetail.Title = "Client Detail"
	clientDetail.BorderStyle.Fg = ui.ColorYellow

	// 创建客户端列表
	clientList := widgets.NewList()
	clientList.Title = "Clients"
	clientList.TextStyle = ui.NewStyle(ui.ColorYellow)
	clientList.WrapText = false
	clientList.SelectedRowStyle = ui.NewStyle(ui.ColorBlack, ui.ColorYellow)

	// 创建基本统计信息区域
	basicStats := widgets.NewParagraph()
	basicStats.Title = "Basic Statistics"
	basicStats.BorderStyle.Fg = ui.ColorYellow

	// 创建请求统计图表
	requestsPlot := widgets.NewPlot()
	requestsPlot.Title = "Requests Per Second"
	requestsPlot.Data = make([][]float64, 1)
	requestsPlot.Data[0] = []float64{0, 0}
	requestsPlot.LineColors = []ui.Color{ui.ColorYellow}
	requestsPlot.BorderStyle.Fg = ui.ColorYellow
	requestsPlot.AxesColor = ui.ColorWhite
	requestsPlot.DrawDirection = widgets.DrawLeft
	requestsPlot.MaxVal = 100

	// 创建错误统计区域
	errorStats := widgets.NewParagraph()
	errorStats.Title = "Error Statistics"
	errorStats.BorderStyle.Fg = ui.ColorRed

	// 创建日志区域
	logView := widgets.NewParagraph()
	logView.Title = "Recent Logs"
	logView.BorderStyle.Fg = ui.ColorBlue

	// 初始化计数器和数据切片
	lastTotal := atomic.LoadUint64(&s.TotalRequests)
	tpsData := []float64{0, 0}

	// 创建事件处理通道
	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// 标记是否在显示详情
	showingDetail := false

	// 设置布局函数
	updateUI := func() {
		if !uiActive {
			return
		}
		// 获取终端大小
		width, height := ui.TerminalDimensions()

		if showingDetail {
			// 显示详情模式
			if len(clientKeys) > 0 {
				selectedClient := clientKeys[clientList.SelectedRow]
				clientDetail.Text = clientDetails[selectedClient]
				// 居中显示详情
				detailWidth := width * 2 / 3
				detailHeight := 8
				startX := (width - detailWidth) / 2
				startY := (height - detailHeight) / 2
				clientDetail.SetRect(startX, startY, startX+detailWidth, startY+detailHeight)
				ui.Render(clientDetail)
			}
		} else {
			// 正常模式
			// 根据客户端数量计算标题高度
			titleHeight := len(s.Clients) + 2 // 标题行 + 客户端行数 + 边框
			title.SetRect(0, 0, width, titleHeight)
			basicStats.SetRect(0, titleHeight, width/2, (height+titleHeight)/2)
			requestsPlot.SetRect(width/2, titleHeight, width, (height+titleHeight)/2)
			errorStats.SetRect(0, (height+titleHeight)/2, width/2, height-3)
			logView.SetRect(width/2, (height+titleHeight)/2, width, height-3)

			// 更新统计数据
			uptime := s.GetUptime()
			totalReqs := atomic.LoadUint64(&s.TotalRequests)
			successReqs := atomic.LoadUint64(&s.SuccessfulRequests)
			failedReqs := atomic.LoadUint64(&s.FailedRequests)
			bytesRecv := atomic.LoadUint64(&s.BytesReceived)
			bytesSent := atomic.LoadUint64(&s.BytesSent)

			// 计算 TPS
			currentTPS := float64(totalReqs-lastTotal) / 1.0
			lastTotal = totalReqs

			// 更新图表数据，确保至少有两个点
			if len(tpsData) < 2 {
				tpsData = []float64{0, currentTPS}
			} else {
				tpsData = append(tpsData, currentTPS)
				if len(tpsData) > 60 {
					tpsData = tpsData[1:]
				}
			}

			// 动态调整最大值
			maxTPS := currentTPS
			for _, v := range tpsData {
				if v > maxTPS {
					maxTPS = v
				}
			}
			requestsPlot.MaxVal = maxTPS * 1.2 // 设置为最大值的 1.2 倍，留出一些空间
			if requestsPlot.MaxVal < 10 {      // 设置最小值，避免图表太扁
				requestsPlot.MaxVal = 10
			}

			requestsPlot.Data[0] = tpsData
			requestsPlot.Title = fmt.Sprintf("Requests Per Second (Current: %.2f)", currentTPS)

			// 计算成功率
			successRate := float64(0)
			if totalReqs > 0 {
				successRate = float64(successReqs) / float64(totalReqs) * 100
			}

			// 更新基本统计信息
			basicStats.Text = fmt.Sprintf(
				"⏱️  Uptime: %s\n"+
					"🔄 Total Requests: %d\n"+
					"✅ Successful: %d\n"+
					"❌ Failed: %d\n"+
					"📥 Bytes Received: %s\n"+
					"📤 Bytes Sent: %s\n"+
					"📊 Success Rate: %.2f%%",
				uptime.Round(time.Second),
				totalReqs,
				successReqs,
				failedReqs,
				formatBytes(bytesRecv),
				formatBytes(bytesSent),
				successRate,
			)

			// 更新错误统计信息
			if failedReqs > 0 {
				var errorText strings.Builder
				errStats := s.GetErrorStats()
				var codes []int
				for code := range errStats {
					codes = append(codes, code)
				}
				sort.Ints(codes)

				for _, code := range codes {
					count := errStats[code]
					percentage := float64(count) / float64(failedReqs) * 100
					errorText.WriteString(fmt.Sprintf("%d %s\nCount: %d (%.2f%%)\n\n",
						code,
						getStatusCodeDesc(code),
						count,
						percentage,
					))
				}
				errorStats.Text = errorText.String()
			} else {
				errorStats.Text = "No errors reported"
			}

			// 更新日志视图
			logView.Text = logger.GetRecentLogs()

			// 渲染所有组件
			ui.Render(title, basicStats, requestsPlot, errorStats, logView)
		}
	}

	// 主事件循环
	logUpdateChan := logger.GetLogUpdateChan()

	for !uiQuit {
		if uiActive {
			select {
			case e := <-uiEvents:
				switch e.ID {
				case "<Enter>":
					showingDetail = !showingDetail
				case "<C-c>":
					ui.Close()
					uiActive = false
					fmt.Println("\nCtrl+C to exit, any other key to return to stats.")
				case "<Resize>":
					updateUI()
				}
				updateUI()
			case <-ticker.C:
				updateUI()
			case <-logUpdateChan:
				updateUI()
			case <-stopChan:
				return
			}
		} else {
			// 普通模式下的事件处理
			select {
			case <-stopChan:
				return
			default:
				// 设置终端为原始模式
				oldState, err := term.MakeRaw(int(syscall.Stdin))
				if err != nil {
					log.Printf("无法设置终端为原始模式: %v", err)
					return
				}

				// 读取一个字符
				buf := make([]byte, 1)
				if n, err := os.Stdin.Read(buf); err == nil && n == 1 {
					// 恢复终端设置
					term.Restore(int(syscall.Stdin), oldState)

					if buf[0] == 3 { // Ctrl+C
						uiQuit = true
					} else {
						// 任意其他键返回 UI 模式
						fmt.Print("\n") // 在切换回 UI 模式前换行，保持输出整洁
						if err := startUI(); err == nil {
							updateUI()
						}
					}
				} else {
					// 恢复终端设置
					term.Restore(int(syscall.Stdin), oldState)
					time.Sleep(100 * time.Millisecond) // 添加短暂延迟，避免 CPU 占用过高
				}
			}
		}
	}
}

func repeat(char rune, count int) []rune {
	result := make([]rune, count)
	for i := range result {
		result[i] = char
	}
	return result
}
