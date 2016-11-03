/*
	支持精准xmlpath 路径查找
	王涛
	2016-11-1 10:11:55
*/
package xmlReader

import (
	"bufio"
	"container/list"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strings"
)

type XmlReader struct {
	//file string
	root *XmlNode
}

// 只为了配置文件用，别和我说xml狗屁规范
type XmlNode struct {
	Name  string
	Value string //如果有多个，那就是你配置文件写错了，我只取第一个
	//Comment    string
	MapAttr    map[string]string
	ChildNodes *list.List
	Parent     *XmlNode
}

func (this *XmlNode) myString(pre int) string {
	prefix := ""
	for i := 0; i < pre; i++ {
		prefix += "  "
	}
	var ret string
	ret = prefix + "<" + this.Name
	for k, v := range this.MapAttr {
		ret += " " + k + "=\"" + v + "\""
	}
	ret += ">\n"
	if len(this.Value) > 0 {
		ret += prefix + "  " + this.Value + "\n"
	}
	if this.ChildNodes != nil {
		for e := this.ChildNodes.Front(); e != nil; e = e.Next() {
			cn, _ := e.Value.(*XmlNode)
			ret += cn.myString(pre + 1)
		}
	}
	ret += prefix + "</" + this.Name + ">\n"
	return ret
}

func (this *XmlNode) String() string {
	return this.myString(0)
}

type XmlNodeList struct {
	NodeList *list.List
}

func (this *XmlNodeList) Add(node *XmlNode) {
	if this.NodeList == nil {
		this.NodeList = list.New()
	}
	this.NodeList.PushBack(node)
}

func (this *XmlNodeList) Get() *XmlNode {
	if this.NodeList == nil {
		return nil
	}
	ret, _ := (this.NodeList.Back().Value).(*XmlNode)
	return ret
}

func (this *XmlNodeList) Remove() bool {
	if this.NodeList == nil || this.NodeList.Len() == 1 {
		return false
	}
	e := this.NodeList.Back()
	node, _ := (e.Value).(*XmlNode)
	//fmt.Println("rem: ", node.Name, ":", unsafe.Pointer(node)) // node.Name not ok
	this.NodeList.Remove(e)
	parent, _ := (this.NodeList.Back().Value).(*XmlNode)
	node.Parent = parent
	if parent.ChildNodes == nil {
		parent.ChildNodes = list.New()
	}
	parent.ChildNodes.PushBack(node)
	return true
}

func NewXmlReader(xmlFile string) (*XmlReader, *error) {

	ret := new(XmlReader)
	if ret.Init(xmlFile) {
		return ret, nil
	}
	return nil, new(error)
}

func (this *XmlReader) Init(file string) bool {
	fReader, err := os.Open(file)
	if err != nil {
		fmt.Errorf("read xml file err:", file)
		return false
	}
	defer fReader.Close() // 此处不能关闭文件，否则后续无法读取数据
	fBufReader := bufio.NewReader(fReader)
	decoder := xml.NewDecoder(fBufReader)
	var nodeList XmlNodeList
	for t, err := decoder.Token(); err == nil; t, err = decoder.Token() {
		switch token := t.(type) {
		// 处理元素开始（标签）
		case xml.StartElement:
			node := new(XmlNode)
			node.Name = token.Name.Local
			//fmt.Println("start node:" + node.Name)
			for _, attr := range token.Attr {
				if node.MapAttr == nil {
					node.MapAttr = make(map[string]string)
				}
				node.MapAttr[attr.Name.Local] = attr.Value
			}
			nodeList.Add(node)
			// 处理元素结束（标签）
		case xml.EndElement:
			nodeList.Remove()
		case xml.CharData:
			node := nodeList.Get()
			if node != nil && len(node.Value) == 0 {
				vet := string([]byte(token))
				le := len(vet)
				for { //简单粗暴点，无性能要求
					vet = strings.Trim(vet, " ")
					vet = strings.Trim(vet, "\n")
					vet = strings.Trim(vet, "\r\n")
					vet = strings.Trim(vet, "\t")
					if len(vet) == le {
						break
					} else {
						le = len(vet)
					}
				}

				node.Value = vet
			}
		/*case xml.Comment:
		node := nodeList.Get()
		if node != nil {
			vet := string([]byte(token))

			node.Comment = vet
		}*/
		default:
		}
	}
	this.root = nodeList.Get()
	//fmt.Println(this.root)
	return true
}

/*获取 xpath 对应的value
bug# 对于<old_pp>
			cao
			<input_path>  a	</input_path>
			fuck
		</old_pp> 中取 "/old_pp"  的值，只取到第一个 cao
	 原因是golang的xml 标准库就这样，没有办法
*/
func (this *XmlReader) GetValue(xpath string) (string, error) {
	node, err := this.GetNode(xpath)
	if err != nil {
		return "", err
	}
	return node.Value, nil
}

/*获取 xpath 节点下所有的一级节点名字
 */
func (this *XmlReader) GetSubNodeName(xpath string) ([]string, error) {
	node, err := this.GetNode(xpath)
	if err != nil {
		return nil, err
	}
	if node.ChildNodes != nil && node.ChildNodes.Len() > 0 {
		ret := make([]string, node.ChildNodes.Len())
		i := 0
		for e := node.ChildNodes.Front(); e != nil; e = e.Next() {
			ret[i] = (e.Value).(*XmlNode).Name
			i++
		}
		return ret, nil
	}
	return nil, nil
}

func (this *XmlReader) GetAttrMap(xpath string) (map[string]string, error) {
	node, err := this.GetNode(xpath)
	if err != nil {
		return nil, err
	}
	if node.MapAttr != nil && len(node.MapAttr) > 0 {
		ret := make(map[string]string)
		for k, v := range node.MapAttr {
			ret[k] = v
		}
		return ret, nil
	}
	return nil, nil
}

func (this *XmlReader) GetNode(xpath string) (*XmlNode, error) {
	if this.root == nil {
		return nil, errors.New("file not be parser correctly")
	}
	tmpL := len(xpath)
	if tmpL < 2 {
		return nil, errors.New("xpath err")
	}
	if xpath[0:1] != "/" {
		return nil, errors.New("xpath must be start with /")
	}
	xpath = xpath[1:]
	if xpath[tmpL-2:] == "/" {
		return nil, errors.New("xpath can not be end with /")
	}
	pathNode := strings.Split(xpath, "/")
	depth := len(pathNode)
	cur := this.root
	for i := 0; i < depth-1; i++ {
		if cur.Name != pathNode[i] || cur.ChildNodes == nil {
			return nil, errors.New("xpath not found:" + xpath)
		}
		e := cur.ChildNodes.Front()
		for ; e != nil; e = e.Next() {
			if (e.Value).(*XmlNode).Name == pathNode[i+1] {
				cur = (e.Value).(*XmlNode)
				break
			}
		}
		if e == nil {
			return nil, errors.New("xpath not found:" + xpath)
		}
	}
	return cur, nil
}

/*
func main() {

	p, _ := NewXmlReader("ppDiff.xml")
	ret, _ := p.GetValue("/config/c1/old_pp")
	fmt.Println(ret, len(ret))
	names, _ := p.GetSubNodeName("/config/c1/old_pp/input_path")
	if names == nil {
		fmt.Println("nil")
	}
	fmt.Println(names)
	attr, _ := p.GetAttrMap("/config/c1/old_pp")
	fmt.Println(attr)
	return

}*/
