import uuid
from pathlib import Path
from typing import AsyncGenerator, Optional, Dict, Any, List
from contextlib import asynccontextmanager

import aioboto3
from botocore.exceptions import ClientError
from fastapi import Depends

from core.config import Settings, get_settings


class MinioClient:
    """异步 MinIO 客户端（基于 S3 协议）"""

    def __init__(self, settings: Settings):
        self.settings = settings
        self.endpoint_url = f"{'https' if settings.MINIO_SECURE else 'http'}://{settings.MINIO_ENDPOINT}"
        self.bucket_name = settings.MINIO_BUCKET
        self.temp_dir = Path(settings.TEMP_DIR)

        # 创建临时目录
        self.temp_dir.mkdir(parents=True, exist_ok=True)

        # 初始化 aioboto3 session
        self.session = aioboto3.Session(
            aws_access_key_id=settings.MINIO_ACCESS_KEY,
            aws_secret_access_key=settings.MINIO_SECRET_KEY,
            region_name=settings.MINIO_REGION,
        )

    @asynccontextmanager
    async def _get_client(self):
        """获取 S3 客户端的上下文管理器"""
        async with self.session.client(
            "s3", endpoint_url=self.endpoint_url, use_ssl=self.settings.MINIO_SECURE
        ) as client:
            yield client

    async def ensure_bucket_exists(self) -> bool:
        """确保存储桶存在，不存在则创建"""
        async with self._get_client() as client:
            try:
                await client.head_bucket(Bucket=self.bucket_name)
                return True
            except ClientError as e:
                if e.response["Error"]["Code"] == "404":
                    # 桶不存在，创建它
                    await client.create_bucket(Bucket=self.bucket_name)
                    return True
                raise

    async def download_to_local(
        self, object_name: str, local_path: Optional[str] = None, chunk_size: int = 8192
    ) -> str:
        """
        将 MinIO 文件下载到本地，并返回绝对路径。
        使用流式读写，内存占用极低。

        Args:
            object_name: MinIO 中的对象名称/路径
            local_path: 本地保存路径，如果为 None 则自动生成
            chunk_size: 每次读取的块大小（字节）

        Returns:
            本地文件的绝对路径

        Raises:
            ClientError: 文件不存在或下载失败
        """
        # 如果没有指定本地路径，自动生成
        if local_path is None:
            file_extension = Path(object_name).suffix
            local_filename = f"{uuid.uuid4()}{file_extension}"
            local_path = self.temp_dir / local_filename
        else:
            local_path = Path(local_path)

        # 确保父目录存在
        local_path.parent.mkdir(parents=True, exist_ok=True)

        async with self._get_client() as client:
            try:
                # 获取对象
                response = await client.get_object(
                    Bucket=self.bucket_name, Key=object_name
                )

                # 读取并写入本地文件
                # aioboto3 的 StreamingBody.read() 不接受 chunk_size 参数
                async with response["Body"] as stream:
                    data = await stream.read()
                    with open(local_path, "wb") as f:
                        f.write(data)

                return str(local_path.absolute())

            except ClientError as e:
                if e.response["Error"]["Code"] == "NoSuchKey":
                    raise FileNotFoundError(f"对象不存在: {object_name}")
                raise

    async def get_file_stream(
        self, object_name: str, chunk_size: int = 8192
    ) -> AsyncGenerator[bytes, None]:
        """
        返回一个异步生成器，按块产出数据。
        适用于大文件的流式处理，无需将整个文件加载到内存。

        Args:
            object_name: MinIO 中的对象名称/路径
            chunk_size: 每次产出的块大小（字节）- 注意: aioboto3 不支持分块读取

        Yields:
            文件数据块（bytes）

        Example:
            async for chunk in minio_client.get_file_stream("path/to/file.pdf"):
                process_chunk(chunk)
        """
        async with self._get_client() as client:
            try:
                response = await client.get_object(
                    Bucket=self.bucket_name, Key=object_name
                )

                # aioboto3 的 StreamingBody.read() 不接受 chunk_size 参数
                # 一次性读取并返回
                async with response["Body"] as stream:
                    data = await stream.read()
                    # 分块产出
                    for i in range(0, len(data), chunk_size):
                        yield data[i:i + chunk_size]

            except ClientError as e:
                if e.response["Error"]["Code"] == "NoSuchKey":
                    raise FileNotFoundError(f"对象不存在: {object_name}")
                raise

    async def upload_file(
        self,
        local_path: str,
        object_name: Optional[str] = None,
        metadata: Optional[Dict[str, str]] = None,
        content_type: Optional[str] = None,
    ) -> str:
        """
        上传文件到 MinIO

        Args:
            local_path: 本地文件路径
            object_name: MinIO 中的对象名称，如果为 None 则使用文件名
            metadata: 自定义元数据
            content_type: 文件类型（MIME type）

        Returns:
            上传后的对象名称
        """
        local_path = Path(local_path)
        if not local_path.exists():
            raise FileNotFoundError(f"本地文件不存在: {local_path}")

        if object_name is None:
            object_name = local_path.name

        extra_args = {}
        if metadata:
            extra_args["Metadata"] = metadata
        if content_type:
            extra_args["ContentType"] = content_type

        async with self._get_client() as client:
            with open(local_path, "rb") as f:
                await client.put_object(
                    Bucket=self.bucket_name, Key=object_name, Body=f, **extra_args
                )

        return object_name

    async def upload_fileobj(
        self,
        file_data: bytes,
        object_name: str,
        metadata: Optional[Dict[str, str]] = None,
        content_type: Optional[str] = None,
    ) -> str:
        """
        上传文件对象（字节数据）到 MinIO

        Args:
            file_data: 文件数据（bytes）
            object_name: MinIO 中的对象名称
            metadata: 自定义元数据
            content_type: 文件类型（MIME type）

        Returns:
            上传后的对象名称
        """
        extra_args = {}
        if metadata:
            extra_args["Metadata"] = metadata
        if content_type:
            extra_args["ContentType"] = content_type

        async with self._get_client() as client:
            await client.put_object(
                Bucket=self.bucket_name, Key=object_name, Body=file_data, **extra_args
            )

        return object_name

    async def delete_file(self, object_name: str) -> bool:
        """
        删除 MinIO 中的文件

        Args:
            object_name: 要删除的对象名称

        Returns:
            是否删除成功
        """
        async with self._get_client() as client:
            try:
                await client.delete_object(Bucket=self.bucket_name, Key=object_name)
                return True
            except ClientError:
                return False

    async def delete_files(self, object_names: List[str]) -> Dict[str, bool]:
        """
        批量删除 MinIO 中的文件

        Args:
            object_names: 要删除的对象名称列表

        Returns:
            {object_name: success} 字典
        """
        async with self._get_client() as client:
            objects = [{"Key": name} for name in object_names]
            response = await client.delete_objects(
                Bucket=self.bucket_name, Delete={"Objects": objects}
            )

            deleted = {obj["Key"]: True for obj in response.get("Deleted", [])}
            errors = {obj["Key"]: False for obj in response.get("Errors", [])}

            return {**deleted, **errors}

    async def file_exists(self, object_name: str) -> bool:
        """
        检查文件是否存在

        Args:
            object_name: 对象名称

        Returns:
            是否存在
        """
        async with self._get_client() as client:
            try:
                await client.head_object(Bucket=self.bucket_name, Key=object_name)
                return True
            except ClientError as e:
                if e.response["Error"]["Code"] == "404":
                    return False
                raise

    async def get_file_metadata(self, object_name: str) -> Optional[Dict[str, Any]]:
        """
        获取文件元数据

        Args:
            object_name: 对象名称

        Returns:
            元数据字典，包含大小、类型、修改时间等
        """
        async with self._get_client() as client:
            try:
                response = await client.head_object(
                    Bucket=self.bucket_name, Key=object_name
                )
                return {
                    "size": response["ContentLength"],
                    "content_type": response.get("ContentType"),
                    "last_modified": response["LastModified"],
                    "etag": response["ETag"],
                    "metadata": response.get("Metadata", {}),
                }
            except ClientError as e:
                if e.response["Error"]["Code"] == "404":
                    return None
                raise

    async def list_files(
        self, prefix: str = "", max_keys: int = 1000
    ) -> List[Dict[str, Any]]:
        """
        列出文件

        Args:
            prefix: 前缀过滤
            max_keys: 最大返回数量

        Returns:
            文件信息列表
        """
        async with self._get_client() as client:
            response = await client.list_objects_v2(
                Bucket=self.bucket_name, Prefix=prefix, MaxKeys=max_keys
            )

            files = []
            for obj in response.get("Contents", []):
                files.append(
                    {
                        "key": obj["Key"],
                        "size": obj["Size"],
                        "last_modified": obj["LastModified"],
                        "etag": obj["ETag"],
                    }
                )

            return files

    async def get_presigned_url(
        self, object_name: str, expires_in: int = 3600, method: str = "get_object"
    ) -> str:
        """
        生成预签名 URL

        Args:
            object_name: 对象名称
            expires_in: 过期时间（秒）
            method: 操作方法 ('get_object' 或 'put_object')

        Returns:
            预签名 URL
        """
        async with self._get_client() as client:
            url = await client.generate_presigned_url(
                ClientMethod=method,
                Params={"Bucket": self.bucket_name, "Key": object_name},
                ExpiresIn=expires_in,
            )
            return url

    async def copy_file(
        self, source_object: str, dest_object: str, source_bucket: Optional[str] = None
    ) -> bool:
        """
        复制文件

        Args:
            source_object: 源对象名称
            dest_object: 目标对象名称
            source_bucket: 源存储桶（默认为当前桶）

        Returns:
            是否复制成功
        """
        if source_bucket is None:
            source_bucket = self.bucket_name

        async with self._get_client() as client:
            try:
                await client.copy_object(
                    Bucket=self.bucket_name,
                    CopySource={"Bucket": source_bucket, "Key": source_object},
                    Key=dest_object,
                )
                return True
            except ClientError:
                return False

    def cleanup_temp_files(self, older_than_hours: int = 24):
        """
        清理临时目录中的旧文件

        Args:
            older_than_hours: 清理超过指定小时数的文件
        """
        import time

        current_time = time.time()
        cutoff_time = current_time - (older_than_hours * 3600)

        for file_path in self.temp_dir.glob("*"):
            if file_path.is_file():
                if file_path.stat().st_mtime < cutoff_time:
                    try:
                        file_path.unlink()
                    except Exception:
                        pass


async def get_minio_client(settings: Settings = Depends(get_settings)) -> MinioClient:
    """
    MinIO 客户端依赖注入构造器

    用法:
        from typing import Annotated
        from fastapi import Depends

        MinioClientDep = Annotated[MinioClient, Depends(get_minio_client)]

        @router.post("/upload")
        async def upload(client: MinioClientDep):
            ...
    """
    client = MinioClient(settings)
    await client.ensure_bucket_exists()
    return client
