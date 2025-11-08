import axios from "axios";
import { CommonResponse } from "@/api/common.ts";
import { getErrorMessage } from "@/utils/base.ts";

export type GithubState = {
  enabled: boolean;
  client_id: string;
  client_secret: string;
};

export type GoogleState = {
  enabled: boolean;
  client_id: string;
  client_secret: string;
};

export type WechatState = {
  enabled: boolean;
  app_id: string;
  app_secret: string;
};

export type QQState = {
  enabled: boolean;
  app_id: string;
  app_secret: string;
};

export type OAuthConfig = {
  github: GithubState;
  google: GoogleState;
  wechat: WechatState;
  qq: QQState;
};

export const initialOAuthConfig: OAuthConfig = {
  github: { enabled: false, client_id: "", client_secret: "" },
  google: { enabled: false, client_id: "", client_secret: "" },
  wechat: { enabled: false, app_id: "", app_secret: "" },
  qq: { enabled: false, app_id: "", app_secret: "" },
};

export type OAuthResponse = CommonResponse & { data?: OAuthConfig };

export async function getOAuthConfig(): Promise<OAuthResponse> {
  try {
    const res = await axios.get("/admin/oauth/view");
    return res.data as OAuthResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e) };
  }
}

export async function setOAuthConfig(cfg: OAuthConfig): Promise<CommonResponse> {
  try {
    const res = await axios.post("/admin/oauth/update", cfg);
    return res.data as CommonResponse;
  } catch (e) {
    return { status: false, error: getErrorMessage(e) };
  }
}
