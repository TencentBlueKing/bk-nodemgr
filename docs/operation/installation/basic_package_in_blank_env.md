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
kubectl -n blueking-nodemgr exec -it bk-nodemgr-file-xxxxxxx-xxxxx -- python3 /bk-nodemgr/support-files/initpackage/init_package.py --auto-select --tenant-id system --set-as-default
```

使用 `--set-as-default` 时，脚本会在全部安装包完成上传和发布后，自动启用各分组的目标包，并将其全部 platform 设为默认版本；命令成功结束即完成初始化，无需再通过页面或 APIGW 手动设置。选择顺序与失败处理见 [初始化包说明](init_packages.md)。
