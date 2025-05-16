package main

import (
	"fmt"
	"protobuf_demo/api"
	"strings"
	"unicode"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	fieldmask_utils "github.com/mennanov/fieldmask-utils"
)

// type Book struct {
// 	Price int64 // ? 区分默认值和0
//   // Price sql.NullInt64 // 自定义结构体
//   // Price *int64 // 指针
// }

// func foo() {
// 	var book Book
// 	// book.Price
// 	book = Book{Price: 0}
// }

func CamelCase(s string) string {
	var result strings.Builder
	capitalizeNext := true
	for _, r := range s {
			if r == '_' {
					capitalizeNext = true
					continue
			}
			if capitalizeNext {
					result.WriteRune(unicode.ToUpper(r))
					capitalizeNext = false
			} else {
					result.WriteRune(r)
			}
	}
	return result.String()
}

// oneof 示例
func oneofDemo() {

	// client
	req1 := &api.NoticeReaderRequest{
		Msg: "today is 2025/5/15!!!!!!!",
		NoticeWay: &api.NoticeReaderRequest_Email{
			Email: "132@qq.com",
		},
	}

	// req2 := &api.NoticeReaderRequest {
	//   Msg: "2025/5/15 is today !!!!",
	//   NoticeWay: &api.NoticeReaderRequest_Phone{
	//     Phone: "13132392312",
	//   },
	// }

	// server
	req := req1
	// 类型断言
	switch v := req.NoticeWay.(type) {
	case *api.NoticeReaderRequest_Email:
		noticeWithEmail(v)
		// fmt.Printf("message is %v\n" , req.Msg)
	case *api.NoticeReaderRequest_Phone:
		noticeWithPhone(v)
	}

}

// 使用google/protobuf/wrappers.proto
// func wrapValueDemo() {
// 	// client
// 	book := api.Book{
// 		Title: "《学习go语言》",
// 		Price: &wrapperspb.Int64Value{Value: 9900},
// 		Memo:  &wrapperspb.StringValue{Value: "好好学习"},
// 	}
// 	// server
// 	if book.GetPrice() == nil { // 没有给price赋值
//     fmt.Println("没有设置price!!!")
// 	} else {
// 		fmt.Println(book.GetPrice().GetValue())
// 	}

// 	if book.GetMemo() != nil {
// 		fmt.Println(book.GetMemo().GetValue())
// 	}
// }

func optionalDemo() {
	// client
	book := api.Book{
		Title: "《学习go语言》",
		Price: proto.Int64(9800),
		Memo:  &wrapperspb.StringValue{Value: "好好学习"},
	}
	if book.Price == nil {
		fmt.Println("没有设置price!!!")
	} else {
		fmt.Println(book.GetPrice())
	}
}

// 使用field_mask实现部分更新实例
func fieldMaskDemo() {
	// client
	paths := []string{"price" , "info.b"} // 更新的字段信息
	req := api.UpdateBookRequest{
		Op: "tang",
		Book: &api.Book{
			Price: proto.Int64(9900),
			Info: &api.Book_Info{
				B: "bbbbb",
			},
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: paths},
	}
	
	// server
	mask , _ := fieldmask_utils.MaskFromProtoFieldMask(req.UpdateMask , CamelCase)
	var bookDst = make(map[string]interface{})
	// 将数据读取到map[string]interface{}
	// fieldmask-utils支持读取到结构体等，更多用法可查看文档。
	fieldmask_utils.StructToMap(mask, req.Book, bookDst)
	// do update with bookDst
	fmt.Printf("bookDst:%#v\n", bookDst)
}

// 发送通知相关的功能函数
func noticeWithEmail(in *api.NoticeReaderRequest_Email) {
	fmt.Printf("notice reader by email:%v\n", in.Email)
}

func noticeWithPhone(in *api.NoticeReaderRequest_Phone) {
	fmt.Printf("notice reader by phone:%v\n", in.Phone)
}

func main() {
	oneofDemo()
	// wrapValueDemo()
	optionalDemo()
	fieldMaskDemo()
}
