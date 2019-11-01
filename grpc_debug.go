package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gogo/protobuf/jsonpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
)

// Reply for test
type Reply struct {
	res []byte
}

var data string
var ctx string
var file string
var method string
var addr string
var tlsCert string
var tlsServerName string

//Reference https://jbrandhorst.com/post/grpc-json/
func init() {
	encoding.RegisterCodec(JSON{
		Marshaler: jsonpb.Marshaler{
			EmitDefaults: true,
			OrigName:     true,
		},
	})

	flag.StringVar(&data, "data", ``, `req data: format:{"name":"liyanbing","age":18}`)
	flag.StringVar(&ctx, "ctx", ``, `ctx data: format:{"user_id":"1"}`)
	flag.StringVar(&file, "file", ``, `data.json`)
	flag.StringVar(&method, "method", "/example.ExampleService/GetName", `/{package}.{Service}/{Method}`)
	flag.StringVar(&addr, "addr", "127.0.0.1:4096", `127.0.0.1:4096`)
	flag.StringVar(&tlsCert, "cert", "", `./cert.pem`)
	flag.StringVar(&tlsServerName, "server_name", "", `hello_server`)
}

// 使用方法：
//  ./bin/grpcdebug -data='{"name":"liyanbing","age":18}' -addr=127.0.0.1:4096 -method=/{package}.{Service}/{Method}
//  ./bin/grpcdebug -file=data.json -addr=127.0.0.1:4096 -method=/{package}.{Service}/{Method}
func main() {
	flag.Parse()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(JSON{}.Name())),
	}

	if tlsCert != "" {
		creds, err := credentials.NewClientTLSFromFile(tlsCert, tlsServerName)
		if err != nil {
			panic(err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	if file != "" {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("ioutil.ReadFile %s failed!err:=%v", file, err)
			os.Exit(1)
		}
		if len(content) > 0 {
			data = string(content)
		}
	}

	c := context.Background()
	if ctx != "" {
		var ctxData map[string]string
		err := json.Unmarshal([]byte(ctx), &ctxData)
		if err != nil {
			os.Exit(1)
		}

		mm := make(map[string]string)
		for k, v := range ctxData {
			mm[k] = v
		}
		md := metadata.New(mm)
		c = metadata.NewOutgoingContext(c, md)
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		panic(err)
	}

	var reply Reply
	fmt.Println("method:", method, "data:", data)
	err = conn.Invoke(c, method, []byte(data), &reply)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(reply.res))
}

// JSON is impl of encoding.Codec
type JSON struct {
	jsonpb.Marshaler
	jsonpb.Unmarshaler
}

// Name is name of JSON
func (j JSON) Name() string {
	return "json"
}

// Marshal is json marshal
func (j JSON) Marshal(v interface{}) (out []byte, err error) {
	return v.([]byte), nil
}

// Unmarshal is json unmarshal
func (j JSON) Unmarshal(data []byte, v interface{}) (err error) {
	v.(*Reply).res = data
	return nil
}
