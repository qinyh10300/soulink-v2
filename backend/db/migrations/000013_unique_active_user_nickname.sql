DO $$
DECLARE
    dup RECORD;
    seq INT;
    new_nick TEXT;
BEGIN
    FOR dup IN
        SELECT id, nickname, LOWER(nickname) AS lnick,
               ROW_NUMBER() OVER (PARTITION BY LOWER(nickname) ORDER BY created_at, id) AS rn
        FROM users
        WHERE deleted_at IS NULL
    LOOP
        IF dup.rn > 1 THEN
            seq := dup.rn;
            LOOP
                new_nick := dup.nickname || '_' || seq;
                EXIT WHEN NOT EXISTS (
                    SELECT 1 FROM users
                    WHERE LOWER(nickname) = LOWER(new_nick)
                      AND deleted_at IS NULL
                );
                seq := seq + 1;
            END LOOP;
            UPDATE users SET nickname = new_nick, updated_at = NOW()
            WHERE id = dup.id;
            RAISE NOTICE 'renamed duplicate nickname: user % "%" -> "%"', dup.id, dup.nickname, new_nick;
        END IF;
    END LOOP;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS ux_users_nickname_active
    ON users (LOWER(nickname))
    WHERE deleted_at IS NULL;
