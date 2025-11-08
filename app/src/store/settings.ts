import { createSlice } from "@reduxjs/toolkit";
import {
  getBooleanMemory,
  getNumberMemory,
  setBooleanMemory,
  setNumberMemory,
} from "@/utils/memory.ts";
import { RootState } from "@/store/index.ts";
import { isMobile } from "@/utils/device";

export const sendKeys = ["Ctrl + Enter", "Enter"];
export const initialSettings = {
  context: true,
  align: false,
  history: 8,
  sender: !isMobile(), // default [mobile: Ctrl + Enter, pc: Enter]
  max_tokens: 2000,
  temperature: 0.6,
  top_p: 1,
  top_k: 5,
  presence_penalty: 0,
  frequency_penalty: 0,
  repetition_penalty: 1,
  show_capabilities: false,
  enable_tools: false,
  enable_json: false,
  parallel_tools: false,
};

export const settingsSlice = createSlice({
  name: "settings",
  initialState: {
    dialog: false,
    context: getBooleanMemory("context", true), // keep context
    align: getBooleanMemory("align", false), // chat textarea align center
    history: getNumberMemory("history_context", 8), // max history context length
    sender: getBooleanMemory("sender", !isMobile()), // sender (false: Ctrl + Enter, true: Enter)
    max_tokens: getNumberMemory("max_tokens", 2000), // max tokens
    temperature: getNumberMemory("temperature", 0.6), // temperature
    top_p: getNumberMemory("top_p", 1), // top_p
    top_k: getNumberMemory("top_k", 5), // top_k
    presence_penalty: getNumberMemory("presence_penalty", 0), // presence_penalty
    frequency_penalty: getNumberMemory("frequency_penalty", 0), // frequency_penalty
    repetition_penalty: getNumberMemory("repetition_penalty", 1), // repetition_penalty
    show_capabilities: getBooleanMemory("show_capabilities", false),
    enable_tools: getBooleanMemory("enable_tools", false),
    enable_json: getBooleanMemory("enable_json", false),
    parallel_tools: getBooleanMemory("parallel_tools", false),
  },
  reducers: {
    toggleDialog: (state) => {
      state.dialog = !state.dialog;
    },
    setDialog: (state, action) => {
      state.dialog = action.payload as boolean;
    },
    openDialog: (state) => {
      state.dialog = true;
    },
    closeDialog: (state) => {
      state.dialog = false;
    },
    setContext: (state, action) => {
      state.context = action.payload as boolean;
      setBooleanMemory("context", action.payload);
    },
    setAlign: (state, action) => {
      state.align = action.payload as boolean;
      setBooleanMemory("align", action.payload);
    },
    setHistory: (state, action) => {
      state.history = action.payload as number;
      setNumberMemory("history_context", action.payload);
    },
    setSender: (state, action) => {
      state.sender = action.payload as boolean;
      setBooleanMemory("sender", action.payload);
    },
    setMaxTokens: (state, action) => {
      state.max_tokens = action.payload as number;
      setNumberMemory("max_tokens", action.payload);
    },
    setTemperature: (state, action) => {
      state.temperature = action.payload as number;
      setNumberMemory("temperature", action.payload);
    },
    setTopP: (state, action) => {
      state.top_p = action.payload as number;
      setNumberMemory("top_p", action.payload);
    },
    setTopK: (state, action) => {
      state.top_k = action.payload as number;
      setNumberMemory("top_k", action.payload);
    },
    setPresencePenalty: (state, action) => {
      state.presence_penalty = action.payload as number;
      setNumberMemory("presence_penalty", action.payload);
    },
    setFrequencyPenalty: (state, action) => {
      state.frequency_penalty = action.payload as number;
      setNumberMemory("frequency_penalty", action.payload);
    },
    setRepetitionPenalty: (state, action) => {
      state.repetition_penalty = action.payload as number;
      setNumberMemory("repetition_penalty", action.payload);
    },
    setShowCapabilities: (state, action) => {
      state.show_capabilities = action.payload as boolean;
      setBooleanMemory("show_capabilities", action.payload);
    },
    toggleShowCapabilities: (state) => {
      state.show_capabilities = !state.show_capabilities;
      setBooleanMemory("show_capabilities", state.show_capabilities);
    },
    setEnableTools: (state, action) => {
      state.enable_tools = action.payload as boolean;
      setBooleanMemory("enable_tools", action.payload);
    },
    setEnableJson: (state, action) => {
      state.enable_json = action.payload as boolean;
      setBooleanMemory("enable_json", action.payload);
    },
    setParallelTools: (state, action) => {
      state.parallel_tools = action.payload as boolean;
      setBooleanMemory("parallel_tools", action.payload);
    },
    resetSettings: (state) => {
      state.context = initialSettings.context;
      state.align = initialSettings.align;
      state.history = initialSettings.history;
      state.sender = initialSettings.sender;
      state.max_tokens = initialSettings.max_tokens;
      state.temperature = initialSettings.temperature;
      state.top_p = initialSettings.top_p;
      state.top_k = initialSettings.top_k;
      state.presence_penalty = initialSettings.presence_penalty;
      state.frequency_penalty = initialSettings.frequency_penalty;
      state.repetition_penalty = initialSettings.repetition_penalty;
      state.show_capabilities = initialSettings.show_capabilities;
      state.enable_tools = initialSettings.enable_tools;
      state.enable_json = initialSettings.enable_json;
      state.parallel_tools = initialSettings.parallel_tools;

      setBooleanMemory("context", initialSettings.context);
      setBooleanMemory("align", initialSettings.align);
      setNumberMemory("history_context", initialSettings.history);
      setBooleanMemory("sender", initialSettings.sender);
      setNumberMemory("max_tokens", initialSettings.max_tokens);
      setNumberMemory("temperature", initialSettings.temperature);
      setNumberMemory("top_p", initialSettings.top_p);
      setNumberMemory("top_k", initialSettings.top_k);
      setNumberMemory("presence_penalty", initialSettings.presence_penalty);
      setNumberMemory("frequency_penalty", initialSettings.frequency_penalty);
      setNumberMemory("repetition_penalty", initialSettings.repetition_penalty);
      setBooleanMemory("show_capabilities", initialSettings.show_capabilities);
      setBooleanMemory("enable_tools", initialSettings.enable_tools);
      setBooleanMemory("enable_json", initialSettings.enable_json);
      setBooleanMemory("parallel_tools", initialSettings.parallel_tools);
    },
  },
});

export const {
  toggleDialog,
  setDialog,
  openDialog,
  closeDialog,
  setContext,
  setAlign,
  setHistory,
  setSender,
  setMaxTokens,
  setTemperature,
  setTopP,
  setTopK,
  setPresencePenalty,
  setFrequencyPenalty,
  setRepetitionPenalty,
  setShowCapabilities,
  toggleShowCapabilities,
  setEnableTools,
  setEnableJson,
  setParallelTools,
  resetSettings,
} = settingsSlice.actions;
export default settingsSlice.reducer;

export const dialogSelector = (state: RootState): boolean =>
  state.settings.dialog;
export const contextSelector = (state: RootState): boolean =>
  state.settings.context;
export const alignSelector = (state: RootState): boolean =>
  state.settings.align;
export const historySelector = (state: RootState): number =>
  state.settings.history;
export const senderSelector = (state: RootState): boolean =>
  state.settings.sender;
export const maxTokensSelector = (state: RootState): number =>
  state.settings.max_tokens;
export const temperatureSelector = (state: RootState): number =>
  state.settings.temperature;
export const topPSelector = (state: RootState): number => state.settings.top_p;
export const topKSelector = (state: RootState): number => state.settings.top_k;
export const presencePenaltySelector = (state: RootState): number =>
  state.settings.presence_penalty;
export const frequencyPenaltySelector = (state: RootState): number =>
  state.settings.frequency_penalty;
export const repetitionPenaltySelector = (state: RootState): number =>
  state.settings.repetition_penalty;
export const showCapabilitiesSelector = (state: RootState): boolean =>
  state.settings.show_capabilities;
export const enableToolsSelector = (state: RootState): boolean =>
  state.settings.enable_tools;
export const enableJsonSelector = (state: RootState): boolean =>
  state.settings.enable_json;
export const parallelToolsSelector = (state: RootState): boolean =>
  state.settings.parallel_tools;
