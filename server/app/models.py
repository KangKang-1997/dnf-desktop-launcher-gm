from typing import List, Optional

from pydantic import BaseModel, Field


class LoginRequest(BaseModel):
    account_name: str = Field(min_length=1, max_length=64)
    password: str = Field(min_length=1, max_length=128)


class RegisterRequest(BaseModel):
    account_name: str = Field(min_length=4, max_length=16)
    password: str = Field(min_length=6, max_length=16)
    qq: str = Field(default="", max_length=32)


class ChangePasswordRequest(BaseModel):
    account_name: str = Field(min_length=4, max_length=16)
    qq: str = Field(default="", max_length=32)
    new_password: str = Field(min_length=6, max_length=16)


class AdminChangePasswordRequest(BaseModel):
    current_password: str = Field(min_length=1, max_length=128)
    new_password: str = Field(min_length=6, max_length=128)


class SetPermissionsRequest(BaseModel):
    permissions: List[str]


class HomeAnnouncementRequest(BaseModel):
    title: str = Field(default="", max_length=80)
    summary: str = Field(default="", max_length=160)
    content: str = Field(default="", max_length=2000)
    poster_url: str = Field(default="", max_length=512)


class HomeSettingsRequest(BaseModel):
    home_title: str = Field(default="", max_length=80)
    home_eyebrow: str = Field(default="", max_length=80)
    client_download_url: str = Field(default="", max_length=512)
    announcements: List[HomeAnnouncementRequest] = Field(default_factory=list)


class CeraQueryRequest(BaseModel):
    uid: Optional[int] = None
    account_name: str = Field(default="", max_length=64)


class AccountResolveRequest(BaseModel):
    uid: Optional[int] = None
    account_name: str = Field(default="", max_length=64)


class MailSendRequest(BaseModel):
    charac_no: int
    message: str = Field(default="", max_length=512)
    item_id: Optional[int] = None
    item_count: int = Field(default=0, ge=0, le=2147483647)
    gold: int = Field(default=0, ge=0, le=2147483647)
    item_type: str = Field(default="", max_length=32)
    item_grade: int = Field(default=0, ge=0, le=4294967295)
    enhancement_level: int = Field(default=0, ge=0, le=31)
    forge_level: int = Field(default=0, ge=0, le=31)
    amplify_option: int = Field(default=0, ge=0, le=4)
    amplify_value: int = Field(default=0, ge=0, le=65535)


class MailMassSendRequest(BaseModel):
    message: str = Field(default="", max_length=512)
    item_id: Optional[int] = None
    item_count: int = Field(default=0, ge=0, le=2147483647)
    gold: int = Field(default=0, ge=0, le=2147483647)
    item_type: str = Field(default="", max_length=32)
    item_grade: int = Field(default=0, ge=0, le=4294967295)
    enhancement_level: int = Field(default=0, ge=0, le=31)
    forge_level: int = Field(default=0, ge=0, le=31)
    amplify_option: int = Field(default=0, ge=0, le=4)
    amplify_value: int = Field(default=0, ge=0, le=65535)


class MailDeleteRequest(BaseModel):
    charac_no: int


class CeraChargeRequest(BaseModel):
    uid: Optional[int] = None
    account_name: str = Field(default="", max_length=64)
    cera_type: str = Field(default="cera", max_length=16)
    action: str = Field(default="add", max_length=16)
    amount: int = Field(ge=0, le=2147483647)


class BanQueryRequest(BaseModel):
    uid: Optional[int] = None
    account_name: str = Field(default="", max_length=64)


class BanSetRequest(BaseModel):
    uid: Optional[int] = None
    account_name: str = Field(default="", max_length=64)
    punish_type: int = Field(default=1)
    days: int = Field(default=365, ge=1, le=3650)
    reason: str = Field(default="", max_length=255)


class CharacterLevelRequest(BaseModel):
    charac_no: int
    level: int = Field(ge=1, le=999)


class CharacterPvpGradeRequest(BaseModel):
    charac_no: int
    pvp_grade: int = Field(ge=0, le=34)


class CharacterPvpPointRequest(BaseModel):
    charac_no: int
    pvp_point: int = Field(ge=0, le=2147483647)


class CharacterJobRequest(BaseModel):
    charac_no: int
    job: int = Field(ge=0)
    grow_type: int = Field(ge=0, le=15)
    wake_flag: int = Field(default=0, ge=0, le=2)
    expert_job: int = Field(default=0, ge=0, le=4)


class CharacterVisibilityRequest(BaseModel):
    charac_no: int


class InventoryQueryRequest(BaseModel):
    charac_no: int
    scope: str = Field(default="inventory", max_length=32)


class InventoryDeleteRequest(BaseModel):
    charac_no: int
    scope: str = Field(default="inventory", max_length=32)
    slot: int = Field(ge=0)


class InventoryClearRequest(BaseModel):
    charac_no: int
    scope: str = Field(default="inventory", max_length=32)


class AvatarQueryRequest(BaseModel):
    charac_no: int


class AvatarHiddenRequest(BaseModel):
    charac_no: int
    ui_ids: List[int]
    hidden_option: int = Field(ge=0, le=255)


class EventAddRequest(BaseModel):
    event_id: int
    parameter1: int = Field(default=1)
    parameter2: int = Field(default=0)


class PvfRefreshRequest(BaseModel):
    pvf_path: str = Field(min_length=1, max_length=512)
    encode: str = Field(default="big5", max_length=32)


class PvfClientMd5Request(BaseModel):
    client_pvf_md5: str = Field(default="", max_length=32)
