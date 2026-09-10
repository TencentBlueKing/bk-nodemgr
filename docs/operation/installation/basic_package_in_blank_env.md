## 初装环境的最简初始化包流程

### 1. 修改helm values
有部分工具的初始化上传可以由k8s job自动完成
```yaml
initPackages:
  enabled: true
  installOnly: true
  ttlSecondsAfterFinished: 86400
  backoffLimit: 3
  parallelism: 1
```
然后正常安装滚动即可

### 2. 手动上传证书/agent/plugin包
等待file pod就绪之后，准备以下包
- 证书包(包含gseca.crt/gse_server.crt等证书的压缩包)
- gse agent包(如gse_agent_ce-v2.1.6-alpha.65.tgz)
- gse server包(如gse_ce-v2.1.6-alpha.65.tgz)
- nodemgr的relay插件包(如bk-nodemgr-relay-v3.0.1-alpha.82.tgz)
- 其他插件包(如bkunifylogbeat-7.7.2-rc.111.tgz)
将他们以如下的形式放在文件夹里
```bash
packages/
├── agent
│   └── gse_agent_ce-v2.1.6-alpha.65.tgz
├── cert
│   └── cert.tgz
├── plugin-v2
│   └── bkunifylogbeat-7.7.2-rc.111.tgz
├── plugin-v3
│   └── bk-nodemgr-relay-v3.0.1-alpha.82.tgz
└── server
    └── gse_ce-v2.1.6-alpha.65.tgz
```
将这些内容拷贝到file pod里的`/bk-nodemgr/file/packages/`文件夹下
```bash
kubectl -n blueking-nodemgr cp -r ./packages/. bk-nodemgr-file-xxxxxxx-xxxxx:/bk-nodemgr/file/packages/
```
执行自动上传脚本，上传包到system租户(若是单租户环境指定为default)
```bash
kubectl -n blueking-nodemgr exec -it bk-nodemgr-file-xxxxxxx-xxxxx -- python3 /bk-nodemgr/support-files/initpackage/init_package.py --auto-select --tenant-id system
```

### 3. 完成上传后，启用安装包，并设置为默认
完成上传之后，可以通过页面的"包管理"界面看到具体的上传内容。
在用户可以使用之前，还必须设置包为"启用"和"指定默认版本"，可以在页面上手动操作，也可以通过APIGW的接口：
```bash
// 启用linux/amd64的agent包
curl -H 'X-Bk-Tenant-Id: system' -H 'X-Bkapi-Authorization: {"bk_app_code":"xxx","bk_app_secret":"xxx","bk_username":"bk_admin"}' -d '{"generation":2,"release_type":"agent","platform":{"os_type":"linux","cpu_arch":"amd64"},"version":"v2.1.6-alpha.65"}' https://bkapi.blueking-example.com/api/bk-nodemgr/prod/api/v3/package/release/agent/enable

// 设置该agent包为默认包
curl -H 'X-Bk-Tenant-Id: system' -H 'X-Bkapi-Authorization: {"bk_app_code":"xxx","bk_app_secret":"xxx","bk_username":"bk_admin"}' -d '{"generation":2,"release_type":"agent","platform":{"os_type":"linux","cpu_arch":"amd64"},"version":"v2.1.6-alpha.65"}' http://bkapi.blueking-example.com/api/bk-nodemgr/prod/api/v3/package/release/agent/set_as_default

// 启用linux/amd64的proxy包
curl -H 'X-Bk-Tenant-Id: system' -H 'X-Bkapi-Authorization: {"bk_app_code":"xxx","bk_app_secret":"xxx","bk_username":"bk_admin"}' -d '{"generation":2,"release_type":"proxy","platform":{"os_type":"linux","cpu_arch":"amd64"},"version":"v2.1.6-alpha.65"}' https://bkapi.blueking-example.com/api/bk-nodemgr/prod/api/v3/package/release/proxy/enable

// 设置该proxy包为默认包
curl -H 'X-Bk-Tenant-Id: system' -H 'X-Bkapi-Authorization: {"bk_app_code":"xxx","bk_app_secret":"xxx","bk_username":"bk_admin"}' -d '{"generation":2,"release_type":"proxy","platform":{"os_type":"linux","cpu_arch":"amd64"},"version":"v2.1.6-alpha.65"}' http://bkapi.blueking-example.com/api/bk-nodemgr/prod/api/v3/package/release/proxy/set_as_default

// 启用linux/amd64的relay插件包
curl -H 'X-Bk-Tenant-Id: system' -H 'X-Bkapi-Authorization: {"bk_app_code":"xxx","bk_app_secret":"xxx","bk_username":"bk_admin"}' -d '{"generation":2,"name": "bk-nodemgr-relay","platform":{"os_type":"linux","cpu_arch":"amd64"},"version":"v3.0.1-alpha.82"}' https://bkapi.blueking-example.com/api/bk-nodemgr/prod/api/v3/package/release/plugin/enable

// 设置该插件包为默认包
curl -H 'X-Bk-Tenant-Id: system' -H 'X-Bkapi-Authorization: {"bk_app_code":"xxx","bk_app_secret":"xxx","bk_username":"bk_admin"}' -d '{"generation":2,"name": "bk-nodemgr-relay","platform":{"os_type":"linux","cpu_arch":"amd64"},"version":"v3.0.1-alpha.82"}' https://bkapi.blueking-example.com/api/bk-nodemgr/prod/api/v3/package/release/plugin/set_as_default
```

