## grpc 调试工具

要使用此功能，确保server启动的时候import了以下代码

```
import _ "github.com/LiYanBing/grpc_debug/grpc_encoding/json"

```
#### 安装 grpc_debug

```
go get github.com/LiYanBing/grpc_debug
```

#### 查看具体参数用法

```
cd $GOPATH/bin
./grpc_debug -h
```

#### 参数说明：

##### --protocol : debug协议方式：

   grpc：当前请求直接通过GRPC方式请求addr对应的地址；
   
   http：本地会启一个http的服务(监听http_addr)，然后可以通过http的方式将请求参数上传
   内部再通过GRPC的方式发送到addr的服务地址上

##### --data : 请求参数,默认值 {}(json格式)


##### --ctx : 通过GRPC发送数据时需要在Context中传递的参数（json格式，且key,value都是string类型）


##### --file : 请求参数文件路径，如果传递了


##### --method: 请求的方法,格式： /{package}.{Service}/{Method} 

   package 表示定义proto文件的package，Service表示rpc方法的Service名称，method表示请求的method

##### --addr : 请求的grpc服务地址ip:port，默认 127.0.0.1:4096

##### --http_addr : 本地http监听的地址 ip:port，默认 127.0.0.1:2048

##### --cert : grpc 服务器证书地址；程序启动的时候传递

##### --server_name : 主机名称；程序启动的时候传递

### 使用示例

#### 直接调用GRPC方式：
##### 直接通过data传递请求参数（json）

```
./bin/grpcdebug 
-data='{"name":"liyanbing","age":18}' 
-addr=127.0.0.1:4096 
-method=/{package}.{Service}/{Method}

```

##### 通过file方式传递请求参数（json）

```
./bin/grpcdebug 
-file=data.json 
-addr=127.0.0.1:4096 
-method=/{package}.{Service}/{Method}
```

### 通过http方式接受参数
启动一个http服务器，通过Postman 或者 curl 发送http请求；内部会转化成GRPC请求
如果在启动的时候传递了 --addr 参数则会在程序启动时与GRPC服务器建立连接

```
// 通过http方式接受参数 
--addr：grpc服务器地址；
--http_addr：http监听地址，参数需要发送到当前地址

./grpc_debug --protocol=http  
--addr=0.0.0.0:4096 
--http_addr=0.0.0.0:2048

// 通过curl发送参数到http上，因为没有传递 grpc_addr 参数；所以内部会一直使用程序启动时建立的GRPC服务

curl 
-H "Content-Type:application/json" 
-X POST 
-d '{"method":"/example.ExampleService/GetName","data":{"name":"liyanbing","age":18}}' 
127.0.0.1:2048
```

启动一个http服务器，通过POSTMan 或者 curl 发送http请求；内部会转化成GRPC请求
如果在启动的时候不传递 --addr 参数，则只会启动http服务器等待接受参数，所以在向http发送数据时需要传递 grpc_addr参数

```
// 通过http方式接受参数 
--http_addr：http监听地址，参数需要发送到当前地址
./grpc_debug 
--protocol=http 
--http_addr=0.0.0.0:2048

// 通过curl发送参数到http上，如果在程序启动时没有传递 
--addr 则在通过http发送参数时必须传递 grpc_addr 参数；一旦建立连接则后续的请求一直会发送到该GRPC服务器上；除非重新传递了grpc_addr参数
curl 
-H "Content-Type:application/json" 
-X POST 
-d '{"method":"/example.ExampleService/GetName","data":{"name":"liyanbing","age":18},"grpc_addr":"127.0.0.1:4096"}' 
127.0.0.1:2048
```

##### http 请求参数

```
{
	"grpc_addr":"127.0.0.1:4096", // grpc地址，如果在程序启动时没有传递 --addr 则在http请求时需要传递该参数与GRPC服务器建立连接；一旦与GRPC服务器建立连接则后续请求可以不需要该参数
	"method":"/example.ExampleService/GetName", // grpc 请求地址
	"data":{ // 发送给 grpc 请求的数据 json格式
		"name":"liyanbing",
		"age":18
	},
	"ctx":{ // 需要传递到Context中的参数，key,value必须全为string
		"uid":"10000"
	}
}
```
tips : 可以安装Postman，通过Postman发送http请求更加直观
