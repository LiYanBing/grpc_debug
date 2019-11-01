## grpc 调试工具

要使用此功能，确保server启动的时候import了以下代码

```
import _ "github.com/LiYanBing/grpc_debug/grpc_encoding/json"
```
1. 安装 grpc_debug
```
go get github.com/LiYanBing/grpc_debug
```
2. 查看具体参数用法
```
./bin/grpc_debug -h

```
参数说明

-data : 请求参数，json格式

-file : 请求参数文件路径

-method: 请求的方法,格式： /{package}.{Service}/{Method} ,  package 表示定义proto文件的package，Service表示rpc方法的Service名称，method表示请求的method

-addr：请求的服务器ip:port，127.0.0.1:4096

3. 请求示例
```
./bin/grpcdebug -data='{"name":"liyanbing",age:18}' -addr=127.0.0.1:4096 -method=/example.ExampleService/GetName
```

tips: 配合 jq 插件，返回结果更直观

```
brew install jq
```