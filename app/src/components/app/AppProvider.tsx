import NavBar from "./NavBar.tsx";
import { ThemeProvider } from "@/components/ThemeProvider.tsx";
import DialogManager from "@/dialogs";
import Broadcast from "@/components/Broadcast.tsx";
import { useEffectAsync } from "@/utils/hook.ts";
import { bindMarket, getApiPlans } from "@/api/v1.ts";
import { useDispatch } from "react-redux";
import {
  stack,
  updateMasks,
  updateSupportModels,
  useMessageActions,
} from "@/store/chat.ts";
import { dispatchSubscriptionData, setTheme } from "@/store/globals.ts";
import { infoEvent } from "@/events/info.ts";
import { setForm } from "@/store/info.ts";
import { themeEvent } from "@/events/theme.ts";
import { useEffect } from "react";
import { getFeatureInfo } from "@/admin/api/feature.ts";
import { setBooleanMemory, setMemory } from "@/utils/memory.ts";
import { activeTheme } from "@/components/ThemeProvider.tsx";

function AppProvider() {
  const dispatch = useDispatch();
  const { receive } = useMessageActions();

  useEffect(() => {
    infoEvent.bind((data) => dispatch(setForm(data)));
    themeEvent.bind((theme) => dispatch(setTheme(theme)));

    stack.setCallback(async (id, message) => {
      await receive(id, message);
    });
  }, []);

  useEffectAsync(async () => {
    updateSupportModels(dispatch, await bindMarket());
    dispatchSubscriptionData(dispatch, await getApiPlans());
    await updateMasks(dispatch);
    try {
      const features = await getFeatureInfo();
      const md = features?.markdown || ({} as any);
      setBooleanMemory("feature_md_highlight", md.highlight ?? true);
      setBooleanMemory("feature_md_math", md.math ?? true);
      setBooleanMemory("feature_md_mermaid", md.mermaid ?? true);
      setBooleanMemory("feature_md_chart", md.chart ?? true);
      const th = features?.theme || ({} as any);
      if (th.enforce && th.site_theme) {
        setMemory("theme", th.site_theme);
        activeTheme(th.site_theme);
        dispatch(setTheme(th.site_theme));
      }
    } catch (e) {}
  }, []);

  return (
    <>
      <Broadcast />
      <NavBar />
      <ThemeProvider />
      <DialogManager />
    </>
  );
}

export default AppProvider;
