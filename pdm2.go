package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	TPL = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="description" content="由tjz101@qq.com提供技术驱动">
	<title>{{.Name}} - Powered by pdm2</title>
	<style>
		.main-content {
			width: 98%;
			padding: 0;
			margin: 0 auto;
		}
		h3 {
			font-size: 13px;
		}
		a {
			text-decoration: none;
		}
		.table {
			width: 100%;
			border-collapse: collapse;
			font-size: 13px;
		}
		.table tr {
			line-height: 25px;
		}
		.table thead tr th {
			background: #DAEEF3;
			font-weight: normal;
			text-align: left;
		}
		.table tbody tr td {
			word-wrap: break-word; 
		}
		.table_name {
			font-size: 13px;
			margin: 15px 0 5px 0;
		}
		.table_name a {
			color: #000000;
		}
		.goback {
			font-size: 13px;
			text-align: center;
			margin: 15px 0 15px 0;
		}
		.goback a {
			color: #0000FF;
		}
		.clear-fix {
			clear: both;
			display: inline-table;
		}
		.divider {
			border-bottom: 1px solid green;
		}
	</style>
</head>
<body>
	<div class="main-content">
		<h2>文件信息：</h2>
		<div class="pdm-info">
			<table class="table" border="1">
				<tbody>
					<tr>
						<td>作者：</td>
						<td>{{.Author}}</td>
					</tr>
					<tr>
						<td>版本：</td>
						<td>{{.Version}}</td>
					</tr>
					<tr>
						<td>文件名：</td>
						<td>{{.FileName}}</td>
					</tr>
					<tr>
						<td>注释：</td>
						<td>{{.Comment}}</td>
					</tr>
					<tr>
						<td>数据库：</td>
						<td>{{.DBMS.Shortcut.Name}}</td>
					</tr>
					<tr>
						<td>转换工具：</td>
						<td><a href="http://pdm2.caffol.com/" target="_blank">pdm2 v0.1.0</a></td>
					</tr>
				</tbody>
			</table>
		</div>
		<h2>表清单：</h2>
		<div class="table-info">
			<table class="table" border="1">
				<thead>
					<tr>
						<th>名称</th>
						<th>代码</th>
						<th>注释</th>
					</tr>
				</thead>
				<tbody>
					{{range .Tables.Table}}
					<tr>
						<td><a href="#{{.Code}}">{{.Name}}</a></td>
						<td><a href="#{{.Code}}" id="list_{{.Code}}">{{.Code}}</a></td>
						<td>{{.Comment}}</td>
					</tr>
					{{end}}
				</tbody>
			</table>
		</div>
		<div class="clear-fix"></div>
		<h2>表结构：</h2>
		<div class="table-content">
			{{range .Tables.Table}}
			<div class="table-info">
				<div class="table_name"><a href="javascript:void(0)" id="{{.Code}}">{{.Name}}（{{.Code}}）</a></div>
				<table class="table" border="1">
					<thead>
						<tr>
							<th width="15%">是否主键</th>
							<th width="15%">名称</th>
							<th width="15%">代码</th>
							<th width="15%">类型</th>
							<th width="10%">允许为空</th>
							<th width="15%">默认值</th>
							<th width="15%">注释</th>
						</tr>
					</thead>
					<tbody>
						{{with $tb := .}}
						{{range $tb.Columns.Column}}
						<tr>
							<td>
							{{with $co := .}}
							{{range $tb.Keys.Key}}
							{{if eq .Id $tb.PrimaryKey.Key.Ref}}
							{{range .KeyColumns.KeyColumn}}
							{{if eq .Ref $co.Id}}是{{end}}
							{{end}}
							{{end}}
							{{end}}
							{{end}}
							</td>
							<td>{{.Name}}</td>
							<td>{{.Code}}</td>
							<td>{{.DataType}}</td>
							<td>{{if eq .ColumnMandatory ""}}是{{end}}</td>
							<td>{{.DefaultValue}}</td>
							<td>{{.Comment}}</td>
						</tr>
						{{end}}
						{{end}}
					</tbody>
				</table>
				<div class="goback"><a href="#list_{{.Code}}">返回</a></div>
			</div>
			{{end}}
		</div>
	</div>
</body>
</html>`
)

type Shortcut struct {
	ObjectId string
	Name     string
	Code     string
}

type DBMS struct {
	Shortcut Shortcut
}

type Column struct {
	Id              string `xml:"Id,attr"`
	ObjectId        string
	Name            string
	Code            string
	DataType        string
	Identity        string
	DefaultValue    string
	Comment         string
	ColumnMandatory string `xml:"Column.Mandatory"`
}

type Columns struct {
	Column []Column
}

type KeyColumn struct {
	Ref string `xml:"Ref,attr"`
}
type KeyColumns struct {
	KeyColumn []KeyColumn `xml:"Column"`
}

type Key struct {
	Id         string `xml:"Id,attr"`
	ObjectId   string
	Name       string
	Code       string
	KeyColumns KeyColumns `xml:"Key.Columns"`
}

type Keys struct {
	Key []Key
}

type PrimaryKey struct {
	Key KeyColumn
}
type Table struct {
	ObjectId   string
	Name       string
	Code       string
	Comment    string
	Columns    Columns
	Keys       Keys
	PrimaryKey PrimaryKey
}

type Tables struct {
	Table []Table
}

type Model struct {
	ObjectId string
	Name     string
	Code     string
	Author   string
	Version  string
	FileName string
	Comment  string
	DBMS     DBMS
	Tables   Tables
}

type Children struct {
	Model Model
}

type RootObject struct {
	Children Children
}

type Result struct {
	RootObject RootObject
}

func checkErr(v ...interface{}) {
	fmt.Print(v...)
	os.Exit(1)
}

func exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func toHtml(model Model, filename string) error {
	t := template.New("html")
	t, err := t.Parse(TPL)
	if err != nil {
		return err
	}
	model.FileName = filename
	wd, err := os.Getwd()
	if err != nil {
		checkErr(err)
	}
	htmlFilename := filepath.Join(wd, filepath.Base(filename)+".html")
	if exist(htmlFilename) {
		fmt.Printf("Target file [%s] already exists, overwrite? (y/n)", htmlFilename)
		var yn string
		fmt.Scanf("%s", &yn)
		if strings.ToUpper(yn) == "N" {
			os.Exit(1)
		}
	}
	file, err := os.OpenFile(htmlFilename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		checkErr(err)
	}
	defer file.Close()
	err = t.Execute(file, model)
	if err != nil {
		return err
	}
	//fmt.Println(htmlFilename)
	return nil
}

func main() {
	var filename string

	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	if filename == "" {
		os.Exit(1)
	}

	buff, err := ioutil.ReadFile(filename)
	if err != nil {
		checkErr(err)
	}

	var result Result
	err = xml.Unmarshal(buff, &result)
	if err != nil {
		checkErr(err)
	}

	err = toHtml(result.RootObject.Children.Model, filename)
	if err != nil {
		checkErr(err)
	}
}