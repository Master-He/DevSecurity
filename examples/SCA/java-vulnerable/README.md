# Java SCA 演示项目

该 POM 只用于本地 PoC，不表示应实际使用其中的漏洞版本。

将整个目录压缩为 ZIP，然后在平台项目中上传。SCA 会在有 Maven 时优先执行：

```text
mvn -B -q -DskipTests -Dstyle.color=never dependency:tree
```

如果 `mvn` 不存在、离线无法解析依赖或命令失败，则回退解析 `pom.xml`。未声明版本或无法解析的属性版本会输出为 `unknown`。

当前离线漏洞样本可识别 Log4j2、SnakeYAML、Commons Collections 和 H2 等组件的漏洞。数据仅用于演示匹配和展示流程，不应作为生产漏洞库。
