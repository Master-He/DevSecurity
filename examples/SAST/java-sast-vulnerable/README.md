# Java SAST 漏洞演示

该目录用于演示 Joern SAST 的三条基础规则。将整个目录压缩为 ZIP 后，在项目中创建 `SAST` 扫描并上传。

预期结果：

- `SAST-JAVA-001` HIGH / CWE-78：调用 `Runtime.getRuntime().exec`
- `SAST-JAVA-002` MEDIUM / CWE-327：使用 MD5
- `SAST-JAVA-003` HIGH / CWE-502：调用 `ObjectInputStream.readObject`

这些规则表示需要人工复核的危险 API 使用，不直接证明漏洞可被远程利用。
