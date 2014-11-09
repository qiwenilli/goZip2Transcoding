/*###############################################################################
#     FileName: goZip2Transcoding.go
#       Author: qiwen<34214399@qq.com>
#     HomePage: http://www.phpwind.net
#      Version: $Id
#   LastChange: 2014-11-09 15:47:04
#         Desc: 
###############################################################################*/
package main

import(
    "archive/zip"
    "fmt"
    "bytes"
    "io/ioutil"
    "bufio"
    "strings"
    "os"
    "flag"
    "path/filepath"
    "github.com/qiniu/iconv"
)

var(
    des_path string
    from_chartset string
    out_charset string
)


func main(){

    _s := flag.String("src","","need is zip file")
    _f := flag.String("f","utf-8","input file charset")
    _t := flag.String("t","gbk","output file charset")
    _h := flag.Bool("h",false,"help")
    flag.Parse()

    if *_s=="" || *_h==true {
        flag.Usage()
        os.Exit(0)
    }

    _src_filed := strings.Split(*_s,".")
    if strings.ToLower(_src_filed[len(_src_filed)-1])!="zip" {
        flag.Usage()
        fmt.Println("error: -src need is zip file")
        os.Exit(0)
    }
    des_path = _src_filed[0]+"_"+*_t
    from_chartset = *_f
    out_charset = *_t

    //
    Unzip(*_s, des_path)
    //
    compress(des_path, des_path+".zip")
}

func Unzip(src, dest string) error {
    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer r.Close()

    iconvO, _ := iconv.Open(out_charset, from_chartset)
    defer iconvO.Close()

    for _, f := range r.File {
        rc, err := f.Open()
        if err != nil {
            continue
            //return err
        }
        defer rc.Close()

        despath := filepath.Join(dest, f.Name)
        if f.FileInfo().IsDir() {
            os.MkdirAll(despath, f.Mode())
        } else {
            //f, err := os.OpenFile( path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode() )
            desf, err := os.Create(despath)
            if err != nil {
                return err
            }

            sourceBuf := bufio.NewReader(rc)
            file_contents,_ := ioutil.ReadAll(sourceBuf)

            //这里把文件的编码格式更换掉
            output_str :=  strings.Replace( string(file_contents),"CHARSET=utf8","CHARSET=gbk",-1)
            //output_str := iconvO.ConvString( string(file_contents) )
            output_str = iconvO.ConvString( output_str )
            if len(output_str)>0 {
                desf.WriteString(output_str)
            }else{
                desf.Write(file_contents)
            }
            fmt.Println( despath, len(output_str), err )

            desf.Close()
        }
    }
    return nil
}


// 参数frm可以是文件或目录，不会给dst添加.zip扩展名
func compress(frm, dst string) error {
    buf := bytes.NewBuffer(make([]byte, 0, 10*1024*1024)) // 创建一个读写缓冲
    myzip := zip.NewWriter(buf)                           // 用压缩器包装该缓冲
    // 用Walk方法来将所有目录下的文件写入zip
    err := filepath.Walk(frm, func(path string, info os.FileInfo, err error) error {
        var file []byte
        if err != nil {
            return filepath.SkipDir
        }
        header, err := zip.FileInfoHeader(info) // 转换为zip格式的文件信息
        if err != nil {
            return filepath.SkipDir
        }
        header.Name, _ = filepath.Rel(filepath.Dir(frm), path)
        if !info.IsDir() {
            // 确定采用的压缩算法（这个是内建注册的deflate）
            header.Method = 8
            file, err = ioutil.ReadFile(path) // 获取文件内容
            if err != nil {
                return filepath.SkipDir
            }
        } else {
            file = nil
        }
        // 上面的部分如果出错都返回filepath.SkipDir
        // 下面的部分如果出错都直接返回该错误
        // 目的是尽可能的压缩目录下的文件，同时保证zip文件格式正确
        w, err := myzip.CreateHeader(header) // 创建一条记录并写入文件信息
        if err != nil {
            return err
        }
        _, err = w.Write(file) // 非目录文件会写入数据，目录不会写入数据
        if err != nil {        // 因为目录的内容可能会修改
            return err         // 最关键的是我不知道咋获得目录文件的内容
        }
        return nil

    })
    if err != nil {
        return err
    }
    myzip.Close()               // 关闭压缩器，让压缩器缓冲中的数据写入buf
    file, err := os.Create(dst) // 建立zip文件
    if err != nil {
        return err
    }
    defer file.Close()
    //
    _, err = buf.WriteTo(file) // 将buf中的数据写入文件
    if err != nil {
        return err
    }
    return nil

}


