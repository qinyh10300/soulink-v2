from .models import User


def assign_pools(users: list[User]) -> dict[str, list[User]]:
    """
    Assign users to matching pools based on Q1 (my gender) and Q2 (desired gender).
    Returns dict with keys 'mf_male', 'mf_female', 'mm', 'ff'.
    "男女均可" in Q2 is treated as "男".
    """
    pools = {"mf_male": [], "mf_female": [], "mm": [], "ff": []}

    for u in users:
        my_gender = u.q(1)
        want_gender = u.q(2)
        if not my_gender or not want_gender:
            continue

        if want_gender == "男女均可":
            want_gender = "男"

        if my_gender == "男" and want_gender == "女":
            pools["mf_male"].append(u)
        elif my_gender == "女" and want_gender == "男":
            pools["mf_female"].append(u)
        elif my_gender == "男" and want_gender == "男":
            pools["mm"].append(u)
        elif my_gender == "女" and want_gender == "女":
            pools["ff"].append(u)

    return pools
