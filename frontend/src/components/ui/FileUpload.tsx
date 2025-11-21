'use client';

import React, { forwardRef, InputHTMLAttributes, useState, useRef } from 'react';
import { cn } from '@/lib/utils';
import { Upload, X, File } from 'lucide-react';
import { uploadFile, UploadProgress, UploadResponse } from '@/services/upload';

interface FileItem {
  file: File;
  progress: number;
  status: 'pending' | 'uploading' | 'success' | 'error';
  error?: string;
  response?: UploadResponse;
}

interface FileUploadProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label?: string;
  helperText?: string;
  maxSize?: number;
  maxFiles?: number;
  acceptedTypes?: string[];
  onUpload?: (files: FileItem[]) => void;
  containerClassName?: string;
  labelClassName?: string;
}

/**
 * FileUpload 组件
 * 支持文件拖拽和进度显示
 */
export const FileUpload = forwardRef<HTMLInputElement, FileUploadProps>(
  (
    {
      className,
      label,
      helperText,
      maxSize = 50 * 1024 * 1024,
      maxFiles = 5,
      acceptedTypes,
      onUpload,
      containerClassName,
      labelClassName,
      disabled,
      ...props
    },
    ref
  ) => {
    const [files, setFiles] = useState<FileItem[]>([]);
    const [isDragging, setIsDragging] = useState(false);
    const inputRef = useRef<HTMLInputElement>(null);

    const handleFiles = async (fileList: FileList) => {
      const newFiles: FileItem[] = [];

      // 检查文件数量
      if (files.length + fileList.length > maxFiles) {
        alert(`Maximum ${maxFiles} files allowed`);
        return;
      }

      // 处理每个文件
      for (let i = 0; i < fileList.length; i++) {
        const file = fileList[i];

        // 检查文件大小
        if (file.size > maxSize) {
          newFiles.push({
            file,
            progress: 0,
            status: 'error',
            error: `File size exceeds ${maxSize / 1024 / 1024}MB limit`,
          });
          continue;
        }

        // 检查文件类型
        if (acceptedTypes && !acceptedTypes.includes(file.type)) {
          newFiles.push({
            file,
            progress: 0,
            status: 'error',
            error: 'File type not accepted',
          });
          continue;
        }

        newFiles.push({
          file,
          progress: 0,
          status: 'pending',
        });
      }

      setFiles(prev => [...prev, ...newFiles]);

      // 上传文件
      for (const fileItem of newFiles) {
        if (fileItem.status === 'pending') {
          await uploadFileItem(fileItem);
        }
      }
    };

    const uploadFileItem = async (fileItem: FileItem) => {
      setFiles(prev =>
        prev.map(item =>
          item === fileItem ? { ...item, status: 'uploading' } : item
        )
      );

      const response = await uploadFile(fileItem.file, {
        onProgress: (progress: UploadProgress) => {
          setFiles(prev =>
            prev.map(item =>
              item === fileItem
                ? { ...item, progress: progress.percentage }
                : item
            )
          );
        },
      });

      setFiles(prev =>
        prev.map(item =>
          item === fileItem
            ? {
              ...item,
              status: response.success ? 'success' : 'error',
              error: response.error,
              response,
            }
            : item
        )
      );

      onUpload?.(files);
    };

    const removeFile = (index: number) => {
      setFiles(prev => prev.filter((_, i) => i !== index));
    };

    return (
      <div className={cn('w-full', containerClassName)}>
        {label && (
          <label className={cn('block text-sm font-medium text-neutral-700 mb-2 dark:text-neutral-300', labelClassName)}>
            {label}
            {props.required && <span className="text-error ml-1">*</span>}
          </label>
        )}

        {/* 拖拽区域 */}
        <div
          onDragOver={(e) => {
            e.preventDefault();
            setIsDragging(true);
          }}
          onDragLeave={() => setIsDragging(false)}
          onDrop={(e) => {
            e.preventDefault();
            setIsDragging(false);
            handleFiles(e.dataTransfer.files);
          }}
          className={cn(
            'border-2 border-dashed rounded-lg p-8 text-center transition-colors cursor-pointer',
            isDragging
              ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/30'
              : 'border-neutral-300 dark:border-neutral-600 hover:border-primary-300 dark:hover:border-primary-700',
            disabled && 'opacity-50 cursor-not-allowed'
          )}
          onClick={() => inputRef.current?.click()}
        >
          <input
            ref={(el) => {
              inputRef.current = el;
              if (typeof ref === 'function') ref(el);
              else if (ref) ref.current = el;
            }}
            type="file"
            multiple
            disabled={disabled}
            onChange={(e) => e.target.files && handleFiles(e.target.files)}
            className="hidden"
            accept={acceptedTypes?.join(',')}
            {...props}
          />

          <Upload className="mx-auto w-8 h-8 text-neutral-400 mb-2" />
          <p className="text-sm font-medium text-neutral-900 dark:text-white">
            Drag and drop files here
          </p>
          <p className="text-xs text-neutral-500 mt-1">
            or click to select files
          </p>
        </div>

        {/* 文件列表 */}
        {files.length > 0 && (
          <div className="mt-4 space-y-2">
            {files.map((fileItem, index) => (
              <div
                key={`${fileItem.file.name}-${index}`}
                className="p-3 rounded-lg border border-neutral-200 dark:border-neutral-700 flex items-center gap-3"
              >
                <File size={18} className="text-neutral-400 flex-shrink-0" />

                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-neutral-900 dark:text-white truncate">
                    {fileItem.file.name}
                  </p>
                  <p className="text-xs text-neutral-500">
                    {(fileItem.file.size / 1024).toFixed(2)} KB
                  </p>

                  {fileItem.status === 'uploading' && (
                    <div className="mt-1 h-1 bg-neutral-200 dark:bg-neutral-700 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-primary-500 transition-all duration-300"
                        style={{ width: `${fileItem.progress}%` }}
                      />
                    </div>
                  )}

                  {fileItem.status === 'error' && (
                    <p className="text-xs text-error mt-1">{fileItem.error}</p>
                  )}
                </div>

                <div className="flex items-center gap-2 flex-shrink-0">
                  {fileItem.status === 'success' && (
                    <span className="text-xs font-medium text-green-600 dark:text-green-400">
                      ✓
                    </span>
                  )}
                  {fileItem.status === 'error' && (
                    <span className="text-xs font-medium text-error">
                      ✕
                    </span>
                  )}
                  <button
                    onClick={() => removeFile(index)}
                    className="p-1 hover:bg-neutral-100 dark:hover:bg-neutral-700 rounded transition"
                  >
                    <X size={16} className="text-neutral-500" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {helperText && (
          <p className="mt-2 text-xs text-neutral-500 dark:text-neutral-400">
            {helperText}
          </p>
        )}
      </div>
    );
  }
);

FileUpload.displayName = 'FileUpload';

