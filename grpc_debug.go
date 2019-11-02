package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

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

var protocol string
var data string
var ctx string
var file string
var method string
var addr string
var tlsCert string
var tlsServerName string
var http_addr string

//Reference https://jbrandhorst.com/post/grpc-json/
func init() {
	encoding.RegisterCodec(JSON{
		Marshaler: jsonpb.Marshaler{
			EmitDefaults: true,
			OrigName:     true,
		},
	})

	flag.StringVar(&protocol, "protocol", "grpc", "grpc or http: default is grpc")
	// grpc args
	flag.StringVar(&data, "data", `{}`, `req data: format:{"name":"liyanbing","age":18}`)
	flag.StringVar(&ctx, "ctx", ``, `ctx data: format:{"user_id":"1"}`)
	flag.StringVar(&file, "file", ``, `request data file path: current value overrides the data value`)
	flag.StringVar(&method, "method", "/example.ExampleService/GetName", `/{package}.{Service}/{Method}`)
	flag.StringVar(&addr, "addr", "127.0.0.1:4096", `grpc addr: default is 127.0.0.1:4096`)
	flag.StringVar(&tlsCert, "cert", "", `./cert.pem`)
	flag.StringVar(&tlsServerName, "server_name", "", `hello_server`)
	// http args
	flag.StringVar(&http_addr, "http_addr", "127.0.0.1:2048", `HTTP local listening address`)
}

func main() {
	flag.Parse()

	protocol = strings.ToUpper(protocol)
	if protocol != "GRPC" && protocol != "HTTP" {
		fmt.Println("invalid protocol")
		return
	}

	// GRPC
	if protocol == "GRPC" {
		GRPCProtocol()
		return
	}

	// HTTP
	HTTPProtocol()
}

func HTTPProtocol() {
	var grpcConn *grpc.ClientConn
	var err error
	if addr != "" {
		grpcConn, err = GetGRPCConn(addr)
		if err != nil {
			fmt.Println("grpc dail Err: ", err)
			return
		}
	}

	err = http.ListenAndServe(http_addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type Request struct {
			GRPCAddr string            `json:"grpc_addr"`
			Method   string            `json:"method"`
			Data     interface{}       `json:"data"`
			Ctx      map[string]string `json:"ctx"`
		}

		var req Request
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			HTTPJSONResponse(w, map[string]interface{}{
				"code": 400,
				"err":  err.Error(),
			})
			return
		}

		// check grpc method
		if req.Method == "" {
			HTTPJSONResponse(w, map[string]interface{}{
				"code": 400,
				"err":  "invalid method",
			})
			return
		}

		// check grpc connection
		if grpcConn == nil && req.GRPCAddr == "" {
			HTTPJSONResponse(w, map[string]interface{}{
				"code": 400,
				"err":  "invalid grpc_addr",
			})
			return
		}

		// 如果传递了grpc地址且跟初始化的地址不一样则重新建立连接
		if req.GRPCAddr != "" && req.GRPCAddr != addr {
			grpcConn, err = GetGRPCConn(req.GRPCAddr)
			if err != nil {
				HTTPJSONResponse(w, map[string]interface{}{
					"code": 400,
					"err":  err.Error(),
				})
				return
			}

			addr = req.GRPCAddr
		}

		// grpc data
		grpcData, err := json.Marshal(req.Data)
		if err != nil {
			HTTPJSONResponse(w, map[string]interface{}{
				"code": 400,
				"err":  err.Error(),
			})
			return
		}

		// grpc context
		c := context.Background()
		if len(req.Ctx) > 0 {
			mm := make(map[string]string)
			for k, v := range req.Ctx {
				mm[k] = v
			}
			md := metadata.New(mm)
			c = metadata.NewOutgoingContext(c, md)
		}

		ret, err := Invoke(c, grpcConn, method, grpcData)
		if err != nil {
			HTTPJSONResponse(w, map[string]interface{}{
				"code": 400,
				"err":  err.Error(),
			})
			return
		}

		HTTPJSONResponse(w, map[string]interface{}{
			"code": 200,
			"data": json.RawMessage(ret),
		})
	}))
	if err != nil {
		log.Fatal(err)
	}
}

// http response of json
func HTTPJSONResponse(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	buf := bytes.NewBuffer(nil)
	_ = json.NewEncoder(buf).Encode(data)
	_, _ = w.Write(buf.Bytes())
}

// grpc
func GRPCProtocol() {
	conn, err := GetGRPCConn(addr)
	if err != nil {
		panic(err)
	}

	_, err = Invoke(GetGRPCContext(), conn, method, GetGRPCData())
	if err != nil {
		fmt.Println("invoke err: ", err)
		return
	}
}

// grpc invoke
func Invoke(ctx context.Context, conn *grpc.ClientConn, method string, data []byte) ([]byte, error) {
	fmt.Println("METHOD: ", method, "DATA: ", string(data))

	var reply Reply
	err := conn.Invoke(ctx, method, data, &reply)
	if err != nil {
		return nil, err
	}

	fmt.Println("RESULT: ", string(reply.res))
	return reply.res, nil
}

// grpc context
func GetGRPCContext() context.Context {
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

	return c
}

// parse grpc request data
func GetGRPCData() []byte {
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

	return []byte(data)
}

// GRPC connection
func GetGRPCConn(addr string) (*grpc.ClientConn, error) {
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

	return grpc.Dial(addr, opts...)
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
