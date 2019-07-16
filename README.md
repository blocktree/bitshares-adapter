# bitshares-adapter

## 项目依赖库

- [openwallet](https://github.com/blocktree/openwallet.git)

## 如何测试

openwtester包下的测试用例已经集成了openwallet钱包体系，创建conf目录，新建bts.ini文件，编辑如下内容：

```ini

#wallet api url
ServerAPI = "https://localhost:8080"
# Cache data file directory, default = "", current directory: ./data
dataDir = ""

```