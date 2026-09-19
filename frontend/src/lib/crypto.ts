import { JSEncrypt } from "jsencrypt";

import { t } from "@/i18n";
import request from "./request";
import { toast } from "./util";

// getPublicKey 获取服务端 RSA 公钥（用于加密密码）
async function getPublicKey(): Promise<string> {
  const res: any = await request.get("/v1/public-key");
  const publicKey = res?.data?.public_key;
  if (!publicKey) {
    throw new Error("empty public key");
  }
  return publicKey;
}

// encryptPassword 使用 RSA 公钥加密密码（后端仅接受密文）。
// 加密失败时不再降级为明文，而是提示错误并中止请求。
export async function encryptPassword(password: string): Promise<string> {
  try {
    const publicKey = await getPublicKey();
    const encryptor = new JSEncrypt();
    encryptor.setPublicKey(publicKey);
    const cipherText = encryptor.encrypt(password);
    if (!cipherText) {
      throw new Error("encrypt password failed");
    }
    return cipherText;
  } catch (err) {
    console.error("[crypto] rsa encrypt password failed.", err);
    toast(t("auth.encryptFailed"), "error");
    throw err;
  }
}
