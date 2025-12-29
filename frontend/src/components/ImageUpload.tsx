/**
 * ImageUpload - drag-and-drop image upload component with S3 presigned URLs.
 */
import { useState, useCallback } from 'react';
import { Upload, X, Image as ImageIcon, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { vehiclesApi } from '@/services/api';

interface UploadedImage {
  id?: number;
  url: string;
  s3_key: string;
  is_primary: boolean;
}

interface ImageUploadProps {
  vehicleId: number;
  images: UploadedImage[];
  onImagesChange: (images: UploadedImage[]) => void;
  maxImages?: number;
}

export function ImageUpload({ vehicleId, images, onImagesChange, maxImages = 20 }: ImageUploadProps) {
  const [uploading, setUploading] = useState(false);
  const [dragOver, setDragOver] = useState(false);

  const uploadFile = async (file: File) => {
    try {
      // Get presigned URL
      const { data } = await vehiclesApi.getUploadUrl(vehicleId, file.name, file.type);
      const { upload_url, s3_key, public_url } = data;

      // Upload to S3
      await fetch(upload_url, {
        method: 'PUT',
        body: file,
        headers: { 'Content-Type': file.type },
      });

      // Register with backend
      const isPrimary = images.length === 0;
      await vehiclesApi.addImage(vehicleId, s3_key, public_url, isPrimary);

      return { url: public_url, s3_key, is_primary: isPrimary };
    } catch (error) {
      console.error('Upload failed:', error);
      throw error;
    }
  };

  const handleFiles = useCallback(async (files: FileList | File[]) => {
    const fileArray = Array.from(files).filter(f => f.type.startsWith('image/'));
    if (fileArray.length === 0) return;

    const remaining = maxImages - images.length;
    const toUpload = fileArray.slice(0, remaining);

    setUploading(true);
    try {
      const uploaded = await Promise.all(toUpload.map(uploadFile));
      onImagesChange([...images, ...uploaded]);
    } catch (error) {
      console.error('Upload error:', error);
    } finally {
      setUploading(false);
    }
  }, [images, maxImages, onImagesChange, vehicleId]);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    handleFiles(e.dataTransfer.files);
  }, [handleFiles]);

  const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      handleFiles(e.target.files);
    }
  };

  const removeImage = (index: number) => {
    const newImages = images.filter((_, i) => i !== index);
    // If we removed the primary, make the first one primary
    if (images[index].is_primary && newImages.length > 0) {
      newImages[0].is_primary = true;
    }
    onImagesChange(newImages);
  };

  const setPrimary = (index: number) => {
    const newImages = images.map((img, i) => ({
      ...img,
      is_primary: i === index,
    }));
    onImagesChange(newImages);
  };

  return (
    <div className="space-y-4">
      {/* Drop zone */}
      <div
        onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
        className={`
          border-2 border-dashed rounded-lg p-8 text-center transition-colors
          ${dragOver ? 'border-primary bg-primary/5' : 'border-muted-foreground/25'}
          ${uploading ? 'pointer-events-none opacity-50' : 'cursor-pointer'}
        `}
        onClick={() => document.getElementById('image-input')?.click()}
      >
        <input
          id="image-input"
          type="file"
          accept="image/*"
          multiple
          className="hidden"
          onChange={handleFileInput}
          disabled={uploading}
        />
        {uploading ? (
          <div className="flex flex-col items-center gap-2">
            <Loader2 className="h-10 w-10 animate-spin text-muted-foreground" />
            <p className="text-muted-foreground">Uploading...</p>
          </div>
        ) : (
          <div className="flex flex-col items-center gap-2">
            <Upload className="h-10 w-10 text-muted-foreground" />
            <p className="text-muted-foreground">
              Drag & drop images here, or click to select
            </p>
            <p className="text-sm text-muted-foreground">
              {images.length} / {maxImages} images
            </p>
          </div>
        )}
      </div>

      {/* Image grid */}
      {images.length > 0 && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {images.map((img, index) => (
            <div
              key={img.s3_key || index}
              className={`
                relative aspect-video rounded-lg overflow-hidden border-2
                ${img.is_primary ? 'border-primary' : 'border-transparent'}
              `}
            >
              <img
                src={img.url}
                alt={`Vehicle image ${index + 1}`}
                className="w-full h-full object-cover"
              />
              
              {/* Primary badge */}
              {img.is_primary && (
                <span className="absolute top-2 left-2 bg-primary text-primary-foreground text-xs px-2 py-1 rounded">
                  Primary
                </span>
              )}

              {/* Actions */}
              <div className="absolute top-2 right-2 flex gap-1">
                {!img.is_primary && (
                  <Button
                    size="icon"
                    variant="secondary"
                    className="h-7 w-7"
                    onClick={() => setPrimary(index)}
                    title="Set as primary"
                  >
                    <ImageIcon className="h-4 w-4" />
                  </Button>
                )}
                <Button
                  size="icon"
                  variant="destructive"
                  className="h-7 w-7"
                  onClick={() => removeImage(index)}
                  title="Remove"
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
