-- +goose Up
CREATE TABLE prompt_industries (
    industry_key TEXT PRIMARY KEY,
    perspective TEXT NOT NULL CHECK (perspective IN ('business', 'personal')),
    title TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE prompt_topics (
    topic_key TEXT PRIMARY KEY,
    industry_key TEXT NOT NULL REFERENCES prompt_industries(industry_key) ON DELETE CASCADE,
    title TEXT NOT NULL,
    prompt_module TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX idx_prompt_topics_context ON prompt_topics (industry_key, is_active, sort_order);

CREATE TABLE prompt_topic_aliases (
    topic_key TEXT NOT NULL REFERENCES prompt_topics(topic_key) ON DELETE CASCADE,
    normalized_phrase TEXT NOT NULL,
    weight INTEGER NOT NULL DEFAULT 1 CHECK (weight > 0),
    is_negative BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (topic_key, normalized_phrase)
);
CREATE INDEX idx_prompt_topic_aliases_exact ON prompt_topic_aliases (normalized_phrase) INCLUDE (topic_key, weight, is_negative);
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_prompt_topic_aliases_trgm ON prompt_topic_aliases USING gin (normalized_phrase gin_trgm_ops);

CREATE TABLE prompt_profiles (
    profile_uuid UUID PRIMARY KEY,
    owner_user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    title TEXT NOT NULL,
    perspective TEXT NOT NULL CHECK (perspective IN ('business', 'personal')),
    industry_key TEXT NOT NULL REFERENCES prompt_industries(industry_key),
    description TEXT NOT NULL DEFAULT '',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX ux_prompt_profiles_default_per_perspective ON prompt_profiles(owner_user_uuid, perspective) WHERE is_default;
CREATE INDEX idx_prompt_profiles_owner ON prompt_profiles(owner_user_uuid, perspective, updated_at DESC);

CREATE TABLE prompt_profile_topics (
    profile_uuid UUID NOT NULL REFERENCES prompt_profiles(profile_uuid) ON DELETE CASCADE,
    topic_key TEXT NOT NULL REFERENCES prompt_topics(topic_key),
    source TEXT NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'recommended', 'auto')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (profile_uuid, topic_key)
);

CREATE TABLE call_prompt_contexts (
    call_uuid UUID PRIMARY KEY REFERENCES calls(call_uuid) ON DELETE CASCADE,
    profile_uuid UUID NULL REFERENCES prompt_profiles(profile_uuid) ON DELETE SET NULL,
    owner_user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    topic_keys JSONB NOT NULL DEFAULT '[]'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE analysis_prompt_snapshots (
    call_uuid UUID PRIMARY KEY REFERENCES calls(call_uuid) ON DELETE CASCADE,
    topics_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO prompt_industries (industry_key, perspective, title, sort_order) VALUES
('sales', 'business', 'Продажи и клиентский сервис', 10),
('health', 'business', 'Медицина и здоровье', 20),
('education', 'business', 'Образование', 30),
('real_estate', 'business', 'Недвижимость', 40),
('finance', 'business', 'Финансы и страхование', 50),
('it', 'business', 'IT и технологии', 60),
('hiring', 'personal', 'Трудоустройство и карьера', 10),
('learning', 'personal', 'Обучение и преподавание', 20),
('services', 'personal', 'Бытовые и профессиональные услуги', 30);

INSERT INTO prompt_topics (topic_key, industry_key, title, prompt_module, sort_order) VALUES
('sales_client_call', 'sales', 'Продажа товара или услуги', 'Проверь выявление потребности, понятность ценности, работу с возражениями, договорённость о следующем шаге и риск потери клиента.', 10),
('sales_support', 'sales', 'Поддержка клиента', 'Проверь, понял ли сотрудник проблему, дал ли проверяемое решение, сроки и владельца следующего действия.', 20),
('health_appointment', 'health', 'Запись к врачу или в клинику', 'Проверь корректность записи: услуга, специалист, дата и время, подготовка, стоимость и действия при изменении записи.', 10),
('education_admission', 'education', 'Поступление в вуз или колледж', 'Проверь полноту консультации об образовательной программе, условиях поступления, документах, сроках, стоимости и следующих действиях абитуриента.', 10),
('education_tutoring', 'education', 'Занятие с репетитором', 'Оцени структуру занятия, ясность объяснений, вовлечение ученика, проверку понимания и конкретное домашнее задание.', 20),
('real_estate_viewing', 'real_estate', 'Аренда или покупка недвижимости', 'Проверь уточнение требований к объекту, прозрачность условий и комиссии, договорённость о просмотре и риски недосказанности.', 10),
('finance_credit', 'finance', 'Кредит или банковский продукт', 'Проверь точность объяснения условий, полной стоимости, ограничений, рисков и дальнейших действий клиента.', 10),
('it_b2b', 'it', 'IT-услуга для бизнеса', 'Проверь выявление бизнес-задачи, текущих процессов, лиц принимающих решение, критериев успеха и следующего шага.', 10),
('hiring_go', 'hiring', 'Собеседование Go-разработчика', 'Оцени компетенции кандидата по Go: конкурентность, контекст, ошибки, тестирование, проектирование API, опыт и ясность ответов.', 10),
('hiring_backend', 'hiring', 'Собеседование backend-разработчика', 'Оцени системное мышление, базы данных, API, надёжность, тестирование, опыт и способность объяснять решения.', 20),
('hiring_manager', 'hiring', 'Собеседование на управленческую позицию', 'Оцени лидерство, постановку задач, обратную связь, работу с конфликтами, метрики и релевантность опыта.', 30),
('learning_language', 'learning', 'Урок иностранного языка', 'Оцени речь преподавателя, баланс практики, исправление ошибок, понятность объяснений, вовлечение ученика и план следующего занятия.', 10),
('services_legal', 'services', 'Юридическая консультация', 'Проверь, были ли уточнены факты и документы, обозначены границы консультации, риски, сроки и следующие действия.', 10);

INSERT INTO prompt_topic_aliases (topic_key, normalized_phrase, weight) VALUES
('hiring_go', 'собеседование go разработчика', 10), ('hiring_go', 'golang разработчик', 8),
('hiring_backend', 'собеседование backend разработчика', 10), ('hiring_backend', 'бэкенд разработчик', 8),
('health_appointment', 'запись к врачу', 10), ('health_appointment', 'запись в клинику', 9),
('education_admission', 'поступление в вуз', 10), ('education_admission', 'абитуриент', 8),
('real_estate_viewing', 'покупка квартиры', 8), ('real_estate_viewing', 'аренда квартиры', 8),
('learning_language', 'урок английского', 10), ('learning_language', 'репетитор английского', 9);

-- +goose Down
DROP TABLE IF EXISTS call_prompt_contexts;
DROP TABLE IF EXISTS analysis_prompt_snapshots;
DROP TABLE IF EXISTS prompt_profile_topics;
DROP TABLE IF EXISTS prompt_profiles;
DROP TABLE IF EXISTS prompt_topic_aliases;
DROP TABLE IF EXISTS prompt_topics;
DROP TABLE IF EXISTS prompt_industries;
