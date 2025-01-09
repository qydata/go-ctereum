# 草田链(go-ctereum)

草田链使用 Go 编程语言编写。它是一个功能强大且灵活的工具，支持开发者和用户运行草田链节点、开发智能合约以及与草田链网络交互。

## 功能特点

- **轻量级和高效**：基于 Go 语言开发，提供性能优化和较低的资源占用。
- **可扩展性**：模块化设计，支持插件式扩展。
- **开发者工具**：内置智能合约开发和调试工具，提供丰富的 API。
- **跨平台支持**：兼容 Windows、macOS 和 Linux。

## 安装

### 通过源码安装

1. 克隆仓库：

   ```bash
   git clone https://github.com/qydata/go-ctereum.git
   cd go-ctereum
   ```

2. 安装依赖并构建：

   ```bash
   make geth
   ```

3. 验证安装：

   ```bash
   ./build/bin/geth version

## 使用

### 启动草田链节点

运行以下命令启动一个草田链主网节点：

```bash
geth
```

您可以使用以下参数自定义节点行为：

- `--networkid`：指定网络 ID。
- `--syncmode`：选择同步模式（`full`、`fast`、`light`）。
- `--datadir`：指定数据目录。

### 连接到控制台

启动节点后，可以连接到交互式 JavaScript 控制台：

```bash
geth attach
```

在控制台中，您可以：
- 查询区块链状态
- 与智能合约交互
- 管理账户

### 管理账户

创建新账户：

```bash
geth account new
```

列出账户：

```bash
geth account list
```

解锁账户：

```bash
geth --unlock <account_address>
```

## 开发者指南

### 使用 JSON-RPC API

go-ctereum 提供了丰富的 JSON-RPC 接口，用于与节点交互。例如：

```bash
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' http://localhost:8545
```

### 部署和调用智能合约

通过 geth 控制台或 Web3.js 等库，开发者可以轻松部署和调用智能合约。例如，在控制台中：

```javascript
> var source = "contract HelloWorld { function sayHello() public pure returns (string memory) { return 'Hello, World!'; } }"
> var compiled = eth.compile.solidity(source)
> var contract = eth.contract(compiled.HelloWorld.info.abiDefinition)
> var instance = contract.new({from: eth.accounts[0], data: compiled.HelloWorld.code, gas: 1000000})
```

## 贡献

欢迎社区贡献！您可以通过以下方式参与：

1. 提交问题 (Issue)：报告错误或建议新功能。
2. 提交代码 (Pull Request)：修复问题或改进功能。

在贡献之前，请阅读 [贡献指南](CONTRIBUTING.md)。

## 资源

- 官方网站：[https://ctblock.cn](https://ctblock.cn)
