我们已经安装了 golang 1.23.4 ，不过需要设置 GOROOT=/usr/local/go 作为环境变量，并且创建了 .gitignore 文件，需要把 server 相关的内容放到 server 目录，同时gitignore可能也要有些修改。

主人命令：
@reame.md 按照说明书整理思路开始开发RelayAPI Server 可以先读取 step.md 中之前完成的步骤，继续开发 ，在全部做完后，把所做的步骤补充到step.md 中。
。执行 go 命令时，先加上 GOROOT=/usr/local/go

已完成的步骤：

1. 创建基本项目结构
   - 创建 server 目录
   - 初始化 Go 模块
   - 创建配置文件

2. 实现核心组件
   - 配置模块 (config)：用于加载和管理服务器配置
   - 加密模块 (crypto)：实现 ECC 加密解密功能
   - 模型层 (models)：定义数据模型和验证逻辑
   - 中间件 (middleware)：实现认证和限流功能
   - 服务层 (services)：实现代理服务功能
   - 处理器 (handlers)：处理 API 请求

3. 集成所有组件
   - 更新主程序，集成所有组件
   - 配置路由和中间件
   - 添加健康检查接口

4. 编写文档
   - 创建 server/README.md
   - 添加编译说明
   - 添加配置说明
   - 添加运行说明
   - 添加 API 使用示例

5. 添加单元测试
   - 配置模块测试：测试配置文件加载和默认值
   - 加密模块测试：测试密钥生成、加密解密和密钥导入导出
   - 令牌模块测试：测试令牌有效性和使用计数
   - 代理服务测试：测试请求转发和错误处理

6. 增强加密功能
   - 添加加密方式配置（AES/ECC）
   - 实现 AES 加密
   - 支持自定义密钥和 IV
   - 重构 ECC 加密实现
   - 创建统一的加密接口

下一步计划：

1. 实现数据库连接和操作
   - 创建数据库连接池
   - 实现令牌的 CRUD 操作
   - 添加数据库迁移功能

2. 完善中间件功能
   - 实现完整的令牌验证逻辑
   - 实现请求频率限制
   - 添加日志记录

3. 添加集成测试
   - 端到端测试
   - 性能测试
   - 负载测试

4. 部署相关
   - 创建 Dockerfile
   - 配置 CI/CD
   - 编写部署文档