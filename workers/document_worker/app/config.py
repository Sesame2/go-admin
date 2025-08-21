import os
import yaml
from typing import Dict, Optional


class ConfigLoader:
    def __init__(self):
        self.settings_path = self.get_settings_folder_path()
        env = os.getenv("ENV", "config").lower()
        conf_file = os.path.join(self.settings_path, f"{env}.yaml")

        if not os.path.exists(conf_file):
            raise FileNotFoundError(f"Configuration file {conf_file} not found.")

        with open(conf_file, "r", encoding="utf-8") as f:
            self.config = yaml.safe_load(f) or {}

    def get_settings_folder_path(self) -> str:
        """
        向上递归查找包含 configs/ 的项目根目录
        """
        current_path = os.path.abspath(os.path.dirname(__file__))

        while True:
            candidate = os.path.join(current_path, "configs")
            if os.path.isdir(candidate):
                return candidate
            parent = os.path.dirname(current_path)
            if parent == current_path:
                raise FileNotFoundError("Could not find 'configs' directory.")
            current_path = parent

    def get_section(self, section: str) -> Dict:
        if section in self.config:
            return self.config[section]
        else:
            raise KeyError(f"Section '{section}' not found.")

    def get(self, section: str, option: str) -> Optional[str]:
        return self.config.get(section, {}).get(option)


class Settings:
    loader = ConfigLoader()
    postgres_host = loader.get("database", "host")
    postgres_username = loader.get("database", "username")
    postgres_password = loader.get("database", "password")

    @classmethod
    def ensure_temp_dir(cls):
        """
        确保下载临时路径存在
        """
        os.makedirs(cls.DOWNLOAD_TEMP_PATH, exist_ok=True)
