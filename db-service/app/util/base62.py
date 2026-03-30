BASE62_ALPHABET = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"


def encode_base62(num: int) -> str:
    if num < 0:
        raise ValueError("Base62 encoding only supports non-negative integers")

    if num == 0:
        return BASE62_ALPHABET[0]

    base = len(BASE62_ALPHABET)
    encoded = []

    while num > 0:
        num, remainder = divmod(num, base)
        encoded.append(BASE62_ALPHABET[remainder])

    return "".join(reversed(encoded))