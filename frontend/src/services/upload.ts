/**
 * 文件上传服务
 * 处理文件上传和进度跟踪
 */

export interface UploadProgress {
  loaded: number;
  total: number;
  percentage: number;
}

export interface UploadResponse {
  success: boolean;
  fileId?: string;
  filename?: string;
  size?: number;
  type?: string;
  error?: string;
}

export type ProgressCallback = (progress: UploadProgress) => void;

/**
 * 上传文件到服务器
 */
export async function uploadFile(
  file: File,
  options?: {
    endpoint?: string;
    onProgress?: ProgressCallback;
    maxSize?: number; // 最大文件大小（字节）
  }
): Promise<UploadResponse> {
  const {
    endpoint = '/api/upload',
    onProgress,
    maxSize = 50 * 1024 * 1024, // 默认 50MB
  } = options || {};

  // 验证文件大小
  if (file.size > maxSize) {
    return {
      success: false,
      error: `File size exceeds maximum allowed size of ${formatBytes(maxSize)}`,
    };
  }

  // 创建 FormData
  const formData = new FormData();
  formData.append('file', file);

  try {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();

      // 监听上传进度
      if (onProgress) {
        xhr.upload.addEventListener('progress', (event) => {
          if (event.lengthComputable) {
            const progress: UploadProgress = {
              loaded: event.loaded,
              total: event.total,
              percentage: Math.round((event.loaded / event.total) * 100),
            };
            onProgress(progress);
          }
        });
      }

      // 监听完成
      xhr.addEventListener('load', () => {
        try {
          if (xhr.status >= 200 && xhr.status < 300) {
            const response = JSON.parse(xhr.responseText);
            resolve({
              success: true,
              ...response,
            });
          } else {
            const errorData = JSON.parse(xhr.responseText);
            resolve({
              success: false,
              error: errorData.error || `Upload failed with status ${xhr.status}`,
            });
          }
        } catch (error) {
          reject(error);
        }
      });

      // 监听错误
      xhr.addEventListener('error', () => {
        reject(new Error('Upload failed'));
      });

      xhr.addEventListener('abort', () => {
        reject(new Error('Upload aborted'));
      });

      // 发送请求
      xhr.open('POST', endpoint);
      xhr.send(formData);
    });
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Unknown error',
    };
  }
}

/**
 * 批量上传文件
 */
export async function uploadFiles(
  files: File[],
  options?: {
    endpoint?: string;
    onProgress?: (index: number, progress: UploadProgress) => void;
    concurrency?: number; // 并发数
    maxSize?: number;
  }
): Promise<UploadResponse[]> {
  const { concurrency = 3, ...uploadOptions } = options || {};

  const results: UploadResponse[] = [];
  const queue = [...files];
  let activeUploads = 0;

  return new Promise((resolve, reject) => {
    const processNext = async () => {
      if (queue.length === 0 && activeUploads === 0) {
        resolve(results);
        return;
      }

      if (queue.length > 0 && activeUploads < concurrency) {
        const file = queue.shift()!;
        const index = results.length;
        activeUploads++;

        try {
          const result = await uploadFile(file, {
            ...uploadOptions,
            onProgress: (progress) => {
              options?.onProgress?.(index, progress);
            },
          });

          results.push(result);
        } catch (error) {
          results.push({
            success: false,
            error: error instanceof Error ? error.message : 'Unknown error',
          });
        }

        activeUploads--;
        processNext();
      }
    };

    // 启动初始任务
    for (let i = 0; i < concurrency && i < files.length; i++) {
      processNext();
    }
  });
}

/**
 * 格式化字节大小
 */
function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return Math.round((bytes / Math.pow(k, i)) * Math.pow(10, dm)) / Math.pow(10, dm) + ' ' + sizes[i];
}

/**
 * 验证文件类型
 */
export function validateFileType(
  file: File,
  allowedTypes?: string[]
): boolean {
  if (!allowedTypes || allowedTypes.length === 0) {
    return true;
  }

  return allowedTypes.some(type => {
    if (type.endsWith('/*')) {
      // 支持通配符，如 'image/*'
      const prefix = type.slice(0, -2);
      return file.type.startsWith(prefix);
    }
    return file.type === type;
  });
}

/**
 * 验证文件扩展名
 */
export function validateFileExtension(
  file: File,
  allowedExtensions?: string[]
): boolean {
  if (!allowedExtensions || allowedExtensions.length === 0) {
    return true;
  }

  const ext = file.name.split('.').pop()?.toLowerCase() || '';
  return allowedExtensions.includes(ext);
}

