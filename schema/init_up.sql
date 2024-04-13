CREATE TABLE features(id BIGSERIAL PRIMARY KEY, info TEXT);
CREATE TABLE tags(id BIGSERIAL PRIMARY KEY, info TEXT);
CREATE TABLE banners(
    id BIGSERIAL,
    title TEXT NOT NULL,
    text TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TIME,
    updated_at TIME,
    is_active BOOLEAN,
    PRIMARY KEY(id)
);
CREATE TABLE banners_tags(
    banner_id BIGINT REFERENCES banners(id) ON DELETE CASCADE,
    feature_id BIGINT REFERENCES features(id) ON DELETE CASCADE,
    tag_id BIGINT REFERENCES tags(id) ON DELETE CASCADE
);
CREATE TABLE users(
    id BIGSERIAL PRIMARY KEY,
    token TEXT NOT NULL unique,
    tag_id BIGINT REFERENCES tags(id),
    is_admin BOOLEAN
);
-- Add Data to tables
--Add tags
INSERT INTO tags (info)
VALUES ('first tag');
INSERT INTO tags (info)
VALUES ('second tag');
INSERT INTO tags (info)
VALUES ('third tag');
-- Add features
INSERT INTO features (info)
VALUES ('first feature');
INSERT INTO features (info)
VALUES ('second feature');
INSERT INTO features (info)
VALUES ('third feature');
-- Add banners
INSERT INTO banners (
        title,
        text,
        url,
        created_at,
        updated_at,
        is_active
    )
VALUES (
        'first_title',
        'first_text',
        'first_url',
        NOW(),
        NOW(),
        true
    );
INSERT INTO banners_tags(banner_id, feature_id, tag_id)
VALUES (1, 1, 1);
--
INSERT INTO banners (
        title,
        text,
        url,
        created_at,
        updated_at,
        is_active
    )
VALUES (
        'second_title',
        'second_text',
        'second_url',
        NOW(),
        NOW(),
        true
    );
INSERT INTO banners_tags(banner_id, feature_id, tag_id)
VALUES (2, 2, 1);
INSERT INTO banners_tags(banner_id, feature_id, tag_id)
VALUES (2, 2, 2);
-- --
INSERT INTO banners (
        title,
        text,
        url,
        created_at,
        updated_at,
        is_active
    )
VALUES (
        'third_title',
        'third_text',
        'third_url',
        NOW(),
        NOW(),
        true
    );
INSERT INTO banners_tags(banner_id, feature_id, tag_id)
VALUES (3, 3, 2);
-- --
INSERT INTO banners (
        title,
        text,
        url,
        created_at,
        updated_at,
        is_active
    )
VALUES (
        'fourth_title',
        'fourth_text',
        'fourth_url',
        NOW(),
        NOW(),
        true
    );
INSERT INTO banners_tags(banner_id, feature_id, tag_id)
VALUES (4, 2, 1);
INSERT INTO banners_tags(banner_id, feature_id, tag_id)
VALUES (4, 2, 2);
INSERT INTO banners_tags(banner_id, feature_id, tag_id)
VALUES (4, 2, 3);
--Add Users
INSERT INTO users(token, tag_id, is_admin)
VALUES ('user_UserFirst', 1, false);
--
INSERT INTO users(token, tag_id, is_admin)
VALUES ('user_UserSecond', 2, false);
--
INSERT INTO users(token, tag_id, is_admin)
VALUES ('admin_AdminFirst', 3, true);