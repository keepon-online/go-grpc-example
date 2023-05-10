package service

import (
	"fmt"
	"github.com/keepon-online/go-grpc-example/gen/hello"
	"io"
	"os"
)

type FileServer struct {
	hello.UnimplementedFileServiceServer
}

func (FileServer) DownLoadFile(request *hello.HelloRequest, stream hello.FileService_DownLoadFileServer) error {
	fmt.Println(request)
	file, err := os.Open("conf/server.crt")
	if err != nil {
		return err
	}
	defer file.Close()

	for {
		buf := make([]byte, 2048)
		_, err = file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		stream.Send(&hello.FileResponse{
			Content: buf,
		})
	}
	return nil
}

func (FileServer) UploadFile(stream hello.FileService_UploadFileServer) error {
	for i := 0; i < 10; i++ {
		response, err := stream.Recv()
		fmt.Println(response, err)
	}
	stream.SendAndClose(&hello.HelloResponse{Message: "完毕了"})
	return nil
}
