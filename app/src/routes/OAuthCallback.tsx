import { useEffect } from "react";
import { useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";
import Loader from "@/components/Loader.tsx";
import { validateToken } from "@/store/auth.ts";

export default function OAuthCallback() {
  const dispatch = useDispatch();
  const navigate = useNavigate();

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const jwt = (params.get("jwt") || "").trim();
    if (!jwt) {
      navigate("/login");
      return;
    }

    validateToken(dispatch as any, jwt, async () => {
      await navigate("/app");
    });
  }, []);

  return (
    <div className={`auth`}>
      <Loader prompt={"登录中"} />
    </div>
  );
}
