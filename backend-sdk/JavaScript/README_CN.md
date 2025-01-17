# RelayAPI JavaScript SDK

RelayAPI JavaScript SDK 是一个用于与 RelayAPI 服务器进行交互的客户端库。它提供了简单的接口来生成 API URL、创建令牌，以及发送各种 API 请求。

## 安装

使用 npm 安装：

```bash
npm install relayapi-sdk
```

## 配置

SDK 需要一个配置对象来初始化。你可以从配置文件（`.rai`）加载配置，或直接传入配置对象。配置格式示例：

```json
{
  "version": "1.0.0",
  "server": {
    "host": "http://localhost",
    "port": 8840,
    "base_path": "/relayapi/"
  },
  "crypto": {
    "method": "aes",
    "aes_key": "your-aes-key",
    "aes_iv_seed": "your-iv-seed"
  }
}
```

## 使用示例

### 基本用法

```javascript
import { RelayAPIClient } from 'relayapi-sdk';
import fs from 'fs/promises';
import { OpenAI } from 'openai';

// 从配置文件加载配置
const configContent = await fs.readFile('config.rai', 'utf-8');
const config = JSON.parse(configContent);

// 创建客户端实例
const client = new RelayAPIClient(config);

// 创建令牌
const token = client.createToken({
    apiKey: 'your-api-key',
    maxCalls: 100,
    expireSeconds: 3600,
    provider: 'openai'
});

// 生成 API URL
const baseUrl = client.generateUrl(token);
console.log('Base URL:', baseUrl);
// 输出示例: http://localhost:8840/relayapi/?token=xxxxx&rai_hash=xxxxx

// 在前端代码中将此 URL 用作 OpenAI API 的基础 URL
const openai = new OpenAI({
    baseURL: baseUrl,
    apiKey: 'not-needed' // 实际的 API key 已包含在 token 中
});
```

### 聊天请求

```javascript
const response = await client.chat({
    messages: [
        { role: 'system', content: 'You are a helpful assistant.' },
        { role: 'user', content: 'What is the capital of France?' }
    ],
    model: 'gpt-3.5-turbo',
    temperature: 0.7,
    maxTokens: 1000,
    token: token
});
```

### 图像生成

```javascript
const response = await client.generateImage({
    prompt: 'A beautiful sunset over Paris',
    model: 'dall-e-3',
    size: '1024x1024',
    quality: 'standard',
    n: 1,
    token: token
});
```

### 嵌入向量生成

```javascript
const response = await client.createEmbedding({
    input: 'The quick brown fox jumps over the lazy dog',
    model: 'text-embedding-ada-002',
    token: token
});
```

### 健康检查

```javascript
const status = await client.healthCheck();
```

### 生成 URL

`generateUrl` 方法用于生成包含令牌和哈希参数的完整 API URL：

```javascript
const url = client.generateUrl(token);  // 使用默认的空端点
const url = client.generateUrl(token, 'chat/completions');  // 指定具体端点
```

参数：
- `token` (string)：加密的令牌字符串
- `endpoint` (string, 可选)：API 端点路径，默认为空字符串

该方法会自动将令牌和配置哈希作为 URL 参数添加。

## API 参考

### RelayAPIClient

#### 构造函数

```javascript
new RelayAPIClient(config)
```

- `config`: 配置文件路径（字符串）或配置对象

#### 方法

##### createToken(options)

创建新的令牌。

- `options.apiKey`: API 密钥
- `options.maxCalls`: 最大调用次数（默认：100）
- `options.expireDays`: 过期天数（默认：1）
- `options.provider`: 提供商（默认：'dashscope'）
- `options.extInfo`: 扩展信息（可选）

##### generateUrl(endpoint, token)

生成 API URL。

- `endpoint`: API 端点路径
- `token`: 令牌字符串

##### chat(options)

发送聊天请求。

- `options.messages`: 消息数组
- `options.model`: 模型名称（默认：'gpt-3.5-turbo'）
- `options.temperature`: 温度值（默认：0.7）
- `options.maxTokens`: 最大令牌数（默认：1000）
- `options.token`: 令牌字符串

##### generateImage(options)

生成图像。

- `options.prompt`: 图像描述
- `options.model`: 模型名称（默认：'dall-e-3'）
- `options.size`: 图像尺寸（默认：'1024x1024'）
- `options.quality`: 图像质量（默认：'standard'）
- `options.n`: 生成数量（默认：1）
- `options.token`: 令牌字符串

##### createEmbedding(options)

生成嵌入向量。

- `options.input`: 输入文本
- `options.model`: 模型名称（默认：'text-embedding-ada-002'）
- `options.token`: 令牌字符串

##### healthCheck()

检查服务器健康状态。

## 错误处理

SDK 中的所有方法都会在发生错误时抛出异常。建议使用 try-catch 块来处理可能的错误：

```javascript
try {
    const response = await client.chat({...});
} catch (error) {
    console.error('Error:', error.message);
}
```

## 示例程序

查看 `examples` 目录中的示例程序，了解更多使用方法：

- `chat.js`: 展示了如何使用 SDK 进行聊天、生成图像和嵌入向量等操作

## 许可证

MIT 