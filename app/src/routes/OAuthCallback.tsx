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
    const bind = (params.get("bind") || "").trim();
    const ok = (params.get("ok") || "").trim();
    if (!jwt) {
      if (bind) {
        // binding flow finished, go back to settings
        navigate("/settings?bind=" + encodeURIComponent(bind) + (ok ? "&ok=" + ok : ""));
      } else {
        navigate("/login");
      }
      return;
    }

    validateToken(dispatch as any, jwt, async () => {
      await navigate("/workspace");
    });
  }, []);

  return (
    <div className={`auth`}>
      <Loader prompt={"登录中"} />
    </div>
  );
}
