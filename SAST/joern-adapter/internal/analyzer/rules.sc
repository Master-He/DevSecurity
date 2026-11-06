import java.nio.file.{Files, Paths}

@main def main(): Unit = {
  importCpg("__CPG_PATH__")

  def clean(value: String): String =
    value
      .replace("\\", "\\\\")
      .replace("\t", "\\t")
      .replace("\r", "\\r")
      .replace("\n", "\\n")

  def location(call: io.shiftleft.codepropertygraph.generated.nodes.Call): String = {
    val file = call.file.name.headOption.getOrElse("<unknown>")
    val line = call.lineNumber.getOrElse(0).toString
    s"${clean(file)}\t${line}\t${clean(call.code)}"
  }

  def line(rule: String, call: io.shiftleft.codepropertygraph.generated.nodes.Call): String =
    s"$rule\t${location(call)}"

  val processCalls =
    cpg.call.methodFullNameExact(
      "java.lang.Runtime.exec:java.lang.Process(java.lang.String)",
      "java.lang.Runtime.exec:java.lang.Process(java.lang.String[])",
      "java.lang.Runtime.exec:java.lang.Process(java.lang.String,java.lang.String[])",
      "java.lang.Runtime.exec:java.lang.Process(java.lang.String[],java.lang.String[])",
      "java.lang.ProcessBuilder.start:java.lang.Process()"
    ).l

  val weakDigest =
    cpg.call.methodFullNameExact(
      "java.security.MessageDigest.getInstance:java.security.MessageDigest(java.lang.String)",
      "javax.crypto.Cipher.getInstance:javax.crypto.Cipher(java.lang.String)",
      "javax.crypto.Cipher.getInstance:javax.crypto.Cipher(java.lang.String,java.lang.String)"
    ).where(_.argument(1).code("(?i).*\\b(MD5|SHA-1|DES|DESEDE|RC4)\\b.*")).l

  val deserialization =
    cpg.call.methodFullNameExact(
      "java.io.ObjectInputStream.readObject:java.lang.Object()"
    ).l

  val lines =
    processCalls.map(line("SAST-JAVA-001", _)) ++
      weakDigest.map(line("SAST-JAVA-002", _)) ++
      deserialization.map(line("SAST-JAVA-003", _))

  Files.writeString(Paths.get("__OUTPUT_PATH__"), lines.mkString("\n"))
  println("SAST_FINDINGS=" + lines.size)
}
