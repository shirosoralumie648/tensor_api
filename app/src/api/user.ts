import axios from "axios";

export type Profile = {
  id: number;
  username: string;
  email: string;
  admin: boolean;
};

export type ProfileResponse = {
  status: boolean;
  error?: string;
  data?: Profile;
  token?: string;
};

export async function getProfile(): Promise<ProfileResponse> {
  const res = await axios.get("/user/profile");
  return res.data as ProfileResponse;
}

export async function updateProfile(patch: Partial<Pick<Profile, "username" | "email">>): Promise<ProfileResponse> {
  const res = await axios.post("/user/profile", patch);
  return res.data as ProfileResponse;
}

export type OAuthBinding = {
  provider: string;
  open_id: string;
  union_id: string;
  created_at: string;
};

export type OAuthBindingsResponse = {
  status: boolean;
  error?: string;
  data?: OAuthBinding[];
};

export async function getOAuthBindings(): Promise<OAuthBindingsResponse> {
  const res = await axios.get("/oauth/bindings");
  return res.data as OAuthBindingsResponse;
}

export async function unbindOAuth(
  provider: string,
  code: string,
): Promise<{ status: boolean; error?: string }>{
  const res = await axios.post(`/oauth/${provider}/unbind`, { code });
  return res.data as { status: boolean; error?: string };
}
