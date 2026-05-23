PLATFORMS = ["line", "whatsapp", "instagram", "telegram", "wechat", "kakao", "twitter", "snapchat", "facebook", "tiktok"]

CONTACT_VERBS = ["add", "text", "dm", "message", "contact", "reach", "find me", "hit me up", "follow"]

def detect_violation(description: str) -> tuple[bool, float]:
    text = description.lower()
    platform_hits = [p for p in PLATFORMS if p in text]
    verb_hits = [v for v in CONTACT_VERBS if v in text]

    if platform_hits and verb_hits:
        confidence = 0.95 if len(platform_hits) + len(verb_hits) > 3 else 0.7
        return True, confidence
    return False, 0.0