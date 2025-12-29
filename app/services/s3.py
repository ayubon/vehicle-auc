"""S3 service for file uploads."""
import os
import uuid
from datetime import datetime
from typing import Optional
import boto3
from botocore.config import Config
from botocore.exceptions import ClientError
import structlog

logger = structlog.get_logger(__name__)


class S3Service:
    """Service for S3 file operations."""
    
    def __init__(self):
        self.bucket = os.environ.get('AWS_S3_BUCKET', 'vehicle-auc-dev')
        self.region = os.environ.get('AWS_REGION', 'us-east-1')
        
        # Check if AWS credentials are configured
        self.enabled = bool(
            os.environ.get('AWS_ACCESS_KEY_ID') and 
            os.environ.get('AWS_SECRET_ACCESS_KEY')
        )
        
        if self.enabled:
            self.client = boto3.client(
                's3',
                region_name=self.region,
                config=Config(signature_version='s3v4')
            )
            logger.info("S3 service initialized", bucket=self.bucket)
        else:
            self.client = None
            logger.warning("S3 credentials not configured, using mock mode")
    
    def generate_upload_url(
        self, 
        vehicle_id: int, 
        filename: str,
        content_type: str = 'image/jpeg',
        expires_in: int = 3600
    ) -> dict:
        """
        Generate a presigned URL for uploading a file.
        
        Args:
            vehicle_id: ID of the vehicle
            filename: Original filename
            content_type: MIME type of the file
            expires_in: URL expiration in seconds
            
        Returns:
            dict with upload_url, s3_key, and public_url
        """
        # Generate unique S3 key
        ext = filename.rsplit('.', 1)[-1].lower() if '.' in filename else 'jpg'
        unique_id = uuid.uuid4().hex[:12]
        timestamp = datetime.utcnow().strftime('%Y%m%d')
        s3_key = f"vehicles/{vehicle_id}/{timestamp}_{unique_id}.{ext}"
        
        if not self.enabled:
            # Return mock URLs for development
            return {
                'upload_url': f'https://{self.bucket}.s3.{self.region}.amazonaws.com/{s3_key}',
                's3_key': s3_key,
                'public_url': f'https://{self.bucket}.s3.{self.region}.amazonaws.com/{s3_key}',
                'mock': True,
            }
        
        try:
            upload_url = self.client.generate_presigned_url(
                'put_object',
                Params={
                    'Bucket': self.bucket,
                    'Key': s3_key,
                    'ContentType': content_type,
                },
                ExpiresIn=expires_in,
            )
            
            public_url = f'https://{self.bucket}.s3.{self.region}.amazonaws.com/{s3_key}'
            
            logger.info("Generated presigned upload URL", s3_key=s3_key)
            
            return {
                'upload_url': upload_url,
                's3_key': s3_key,
                'public_url': public_url,
            }
            
        except ClientError as e:
            logger.error("Failed to generate presigned URL", error=str(e))
            raise
    
    def delete_file(self, s3_key: str) -> bool:
        """Delete a file from S3."""
        if not self.enabled:
            logger.info("Mock delete", s3_key=s3_key)
            return True
        
        try:
            self.client.delete_object(Bucket=self.bucket, Key=s3_key)
            logger.info("Deleted file from S3", s3_key=s3_key)
            return True
        except ClientError as e:
            logger.error("Failed to delete file", s3_key=s3_key, error=str(e))
            return False
    
    def generate_download_url(self, s3_key: str, expires_in: int = 3600) -> Optional[str]:
        """Generate a presigned URL for downloading a file."""
        if not self.enabled:
            return f'https://{self.bucket}.s3.{self.region}.amazonaws.com/{s3_key}'
        
        try:
            url = self.client.generate_presigned_url(
                'get_object',
                Params={'Bucket': self.bucket, 'Key': s3_key},
                ExpiresIn=expires_in,
            )
            return url
        except ClientError as e:
            logger.error("Failed to generate download URL", s3_key=s3_key, error=str(e))
            return None


# Singleton instance
s3_service = S3Service()
