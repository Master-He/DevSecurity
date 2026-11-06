import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;

public class SafeSastDemo {
    public byte[] digest(byte[] data) throws Exception {
        return MessageDigest.getInstance("SHA-256").digest(data);
    }

    public byte[] digestText(String value) throws Exception {
        return digest(value.getBytes(StandardCharsets.UTF_8));
    }
}
