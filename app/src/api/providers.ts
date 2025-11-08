import axios from "axios";
import { getV1Path } from "@/api/v1.ts";

export type ProviderFeatureSet = {
  text: boolean;
  vision: boolean;
  tools: boolean;
  images: boolean;
  audio: boolean;
  video: boolean;
  embeddings: boolean;
  context?: number;
  json?: boolean;
  parallel_tools?: boolean;
};

export type ProviderCapability = {
  provider: string;
  features: ProviderFeatureSet;
};

export type ProviderCapabilitiesMap = Record<string, ProviderCapability>;

const fallback: ProviderCapabilitiesMap = {
  OpenAI: {
    provider: "OpenAI",
    features: {
      text: true,
      vision: true,
      tools: true,
      images: true,
      audio: true,
      video: false,
      embeddings: true,
      context: 128000,
      json: true,
      parallel_tools: true,
    },
  },
  Anthropic: {
    provider: "Anthropic",
    features: {
      text: true,
      vision: true,
      tools: true,
      images: false,
      audio: false,
      video: false,
      embeddings: false,
      context: 200000,
      json: false,
      parallel_tools: false,
    },
  },
  Google: {
    provider: "Google",
    features: {
      text: true,
      vision: true,
      tools: true,
      images: true,
      audio: true,
      video: false,
      embeddings: true,
      context: 100000,
      json: true,
      parallel_tools: false,
    },
  },
  DeepSeek: {
    provider: "DeepSeek",
    features: {
      text: true,
      vision: false,
      tools: true,
      images: false,
      audio: false,
      video: false,
      embeddings: false,
      context: 32000,
      json: false,
      parallel_tools: false,
    },
  },
  "Alibaba Qwen": {
    provider: "Alibaba Qwen",
    features: {
      text: true,
      vision: true,
      tools: true,
      images: true,
      audio: true,
      video: false,
      embeddings: true,
      context: 128000,
    },
  },
  "Zhipu GLM": {
    provider: "Zhipu GLM",
    features: {
      text: true,
      vision: true,
      tools: true,
      images: false,
      audio: false,
      video: false,
      embeddings: true,
      context: 128000,
      json: false,
      parallel_tools: false,
    },
  },
  "Meta Llama": {
    provider: "Meta Llama",
    features: {
      text: true,
      vision: true,
      tools: true,
      images: false,
      audio: false,
      video: false,
      embeddings: true,
      context: 128000,
    },
  },
  Mistral: {
    provider: "Mistral",
    features: {
      text: true,
      vision: false,
      tools: true,
      images: false,
      audio: false,
      video: false,
      embeddings: true,
      context: 32000,
      json: true,
      parallel_tools: true,
    },
  },
  Moonshot: {
    provider: "Moonshot",
    features: {
      text: true,
      vision: true,
      tools: true,
      images: false,
      audio: false,
      video: false,
      embeddings: false,
      context: 128000,
      json: false,
      parallel_tools: false,
    },
  },
  MiniMax: {
    provider: "MiniMax",
    features: {
      text: true,
      vision: true,
      tools: true,
      images: false,
      audio: false,
      video: false,
      embeddings: false,
      context: 64000,
      json: false,
      parallel_tools: false,
    },
  },
  "Baidu ERNIE": {
    provider: "Baidu ERNIE",
    features: {
      text: true,
      vision: true,
      tools: true,
      images: true,
      audio: true,
      video: false,
      embeddings: true,
      context: 32000,
      json: false,
      parallel_tools: false,
    },
  },
  "Tencent Hunyuan": {
    provider: "Tencent Hunyuan",
    features: {
      text: true,
      vision: true,
      tools: true,
      images: true,
      audio: true,
      video: false,
      embeddings: true,
      context: 32000,
    },
  },
  "ByteDance Doubao": {
    provider: "ByteDance Doubao",
    features: {
      text: true,
      vision: true,
      tools: true,
      images: true,
      audio: true,
      video: false,
      embeddings: true,
      context: 64000,
      json: false,
      parallel_tools: false,
    },
  },
  Ollama: {
    provider: "Ollama",
    features: {
      text: true,
      vision: true,
      tools: false,
      images: false,
      audio: false,
      video: false,
      embeddings: true,
      context: 8192,
      json: false,
      parallel_tools: false,
    },
  },
  Other: {
    provider: "Other",
    features: {
      text: true,
      vision: false,
      tools: false,
      images: false,
      audio: false,
      video: false,
      embeddings: false,
      json: false,
      parallel_tools: false,
    },
  },
};

export async function getProviderCapabilities(): Promise<ProviderCapabilitiesMap> {
  try {
    const res = await axios.get(getV1Path("/v1/providers"));
    const data = (res.data || {}) as ProviderCapabilitiesMap;
    const valid = data && typeof data === "object";
    return valid ? data : fallback;
  } catch (e) {
    return fallback;
  }
}

export function getFallbackProviderCapabilities(): ProviderCapabilitiesMap {
  return fallback;
}
