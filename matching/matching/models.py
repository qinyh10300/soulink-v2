from dataclasses import dataclass, field


@dataclass
class User:
    user_id: int
    email: str
    nickname: str
    department: str
    grade: str
    completed: bool
    participating: bool
    answers: dict = field(default_factory=dict)
    importance: dict = field(default_factory=dict)

    def q(self, num):
        return self.answers.get(num)

    def imp(self, num):
        return self.importance.get(num, False)
