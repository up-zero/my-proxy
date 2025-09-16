# MyProxy

> 支持 TCP、UDP 等协议的代理转发，适用于绝大多数网络环境。提供了命令行、WebUI 两种配置代理的方法，极大地简化了代理配置的步骤。

## WebUI 模式

+ 登录

+ 代理管理

## 命令模式

+ 启动服务

```bash
# 默认服务端口 12312
./my-proxy serve
# 指定服务端口
./my-proxy serve -p 12312
```

+ 代理状态

```bash
# 默认查看所有代理的状态
my-proxy status
# 查看指定代理的状态
my-proxy status <name>
```

![cli_status.png](./assets/cli_status.png)

+ 启动代理

```bash
my-proxy start <name>
```

+ 停止代理

```bash
my-proxy stop <name>
```

+ 重启代理

```bash
my-proxy restart <name>
```

+ 创建代理

```bash
my-proxy create <name>
```

![cli_create.png](./assets/cli_create.png)

+ 编辑代理

```bash
my-proxy edit <name>
```

+ 删除代理

```bash
my-proxy del <name>
```
