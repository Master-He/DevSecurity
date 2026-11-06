import java.io.InputStream;
import java.io.ObjectInputStream;
import java.security.MessageDigest;

public class VulnerableSastDemo {
    public void executeExternalCommand(String command) throws Exception {
        Runtime.getRuntime().exec(command);
    }

    public byte[] digestWithMd5(byte[] data) throws Exception {
        return MessageDigest.getInstance("MD5").digest(data);
    }

    public Object deserializeUntrusted(InputStream input) throws Exception {
        return new ObjectInputStream(input).readObject();
    }
}
