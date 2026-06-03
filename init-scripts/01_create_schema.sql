-- =========================================================
-- Seller Recommendation Platform
-- PostgreSQL create script
-- =========================================================

-- ---------------------------------------------------------
-- Таблица регионов
-- Справочник регионов продавцов и заказов
-- ---------------------------------------------------------
create table regions (
    id bigserial primary key, -- Внутренний идентификатор региона
    code varchar(50) not null unique, -- Уникальный код региона
    name varchar(255) not null, -- Человекочитаемое название региона
    is_active boolean not null default true -- Признак активности региона
);

comment on table regions is 'Справочник регионов';
comment on column regions.id is 'PK региона';
comment on column regions.code is 'Уникальный код региона';
comment on column regions.name is 'Название региона';
comment on column regions.is_active is 'Признак активности региона';

-- ---------------------------------------------------------
-- Таблица категорий
-- Справочник категорий товаров
-- ---------------------------------------------------------
create table categories (
    id bigserial primary key, -- Внутренний идентификатор категории
    external_category_id varchar(100) unique, -- Идентификатор категории во внешней системе
    name varchar(255) not null, -- Название категории
    parent_id bigint references categories(id), -- Ссылка на родительскую категорию
    is_active boolean not null default true -- Признак активности категории
);

comment on table categories is 'Справочник категорий товаров';
comment on column categories.id is 'PK категории';
comment on column categories.external_category_id is 'ID категории во внешней системе';
comment on column categories.name is 'Название категории';
comment on column categories.parent_id is 'Родительская категория';
comment on column categories.is_active is 'Признак активности категории';

-- ---------------------------------------------------------
-- Таблица продавцов
-- Главная сущность системы
-- ---------------------------------------------------------
create table sellers (
    id bigserial primary key, -- Внутренний идентификатор продавца
    external_seller_id varchar(100) not null unique, -- Идентификатор продавца во внешней системе
    seller_name varchar(255) not null, -- Название продавца
    seller_type varchar(50) not null, -- Тип продавца: company / ip / self_employed
    status varchar(50) not null, -- Текущий статус продавца
    lifecycle_stage varchar(50) not null, -- Стадия жизненного цикла продавца
    registration_at timestamp with time zone, -- Дата регистрации продавца
    last_login_at timestamp with time zone, -- Дата последнего входа в систему
    home_region_id bigint references regions(id), -- Основной регион продавца
    created_at timestamp with time zone not null default now(), -- Дата создания записи
    updated_at timestamp with time zone not null default now(), -- Дата обновления записи

    constraint chk_sellers_status check (
        status in ('active', 'inactive', 'blocked', 'pending')
    ), -- Ограничение допустимых статусов

    constraint chk_sellers_lifecycle_stage check (
        lifecycle_stage in (
            'registered',
            'catalog_setup',
            'first_sales',
            'active',
            'growth',
            'stagnation',
            'churn_risk',
            'reactivated'
        )
    ) -- Ограничение допустимых стадий жизненного цикла
);

comment on table sellers is 'Продавцы платформы';
comment on column sellers.id is 'PK продавца';
comment on column sellers.external_seller_id is 'ID продавца во внешней системе';
comment on column sellers.seller_name is 'Название продавца';
comment on column sellers.seller_type is 'Тип продавца';
comment on column sellers.status is 'Текущий статус продавца';
comment on column sellers.lifecycle_stage is 'Стадия жизненного цикла продавца';
comment on column sellers.registration_at is 'Дата регистрации';
comment on column sellers.last_login_at is 'Последний вход';
comment on column sellers.home_region_id is 'Основной регион продавца';
comment on column sellers.created_at is 'Дата создания записи';
comment on column sellers.updated_at is 'Дата обновления записи';

-- ---------------------------------------------------------
-- Таблица моделей продаж продавца
-- Какие модели доступны или подключены продавцу
-- ---------------------------------------------------------
create table seller_models (
    id bigserial primary key, -- Внутренний идентификатор записи
    seller_id bigint not null references sellers(id) on delete cascade, -- Ссылка на продавца
    model_code varchar(50) not null, -- Код модели продаж
    status varchar(50) not null, -- Статус модели для продавца
    enabled_at timestamp with time zone, -- Дата включения модели
    disabled_at timestamp with time zone, -- Дата отключения модели
    created_at timestamp with time zone not null default now(), -- Дата создания записи
    updated_at timestamp with time zone not null default now(), -- Дата обновления записи

    constraint chk_seller_models_status check (
        status in ('enabled', 'disabled', 'recommended', 'unavailable')
    ), -- Ограничение допустимых статусов модели

    constraint uq_seller_models unique (seller_id, model_code) -- Одна запись на модель у продавца
);

comment on table seller_models is 'Модели продаж продавца';
comment on column seller_models.id is 'PK записи';
comment on column seller_models.seller_id is 'FK на продавца';
comment on column seller_models.model_code is 'Код модели продаж';
comment on column seller_models.status is 'Статус модели';
comment on column seller_models.enabled_at is 'Дата включения модели';
comment on column seller_models.disabled_at is 'Дата отключения модели';
comment on column seller_models.created_at is 'Дата создания записи';
comment on column seller_models.updated_at is 'Дата обновления записи';

-- ---------------------------------------------------------
-- Таблица товаров продавца
-- Загружается из внешней системы
-- ---------------------------------------------------------
create table seller_products (
    id bigserial primary key, -- Внутренний идентификатор товара
    seller_id bigint not null references sellers(id) on delete cascade, -- Ссылка на продавца
    external_product_id varchar(100) not null, -- Идентификатор товара во внешней системе
    sku varchar(100), -- SKU товара
    product_name varchar(500) not null, -- Название товара
    category_id bigint references categories(id), -- Категория товара
    status varchar(50) not null, -- Статус товара
    price numeric(14,2) not null default 0, -- Текущая цена товара
    stock_qty integer not null default 0, -- Текущий остаток товара
    published_at timestamp with time zone, -- Дата публикации товара
    unpublished_at timestamp with time zone, -- Дата снятия товара
    created_at timestamp with time zone not null default now(), -- Дата создания записи
    updated_at timestamp with time zone not null default now(), -- Дата обновления записи
    loaded_at timestamp with time zone not null default now(), -- Дата загрузки из внешней системы

    constraint chk_seller_products_status check (
        status in ('draft', 'active', 'inactive', 'archived')
    ), -- Ограничение допустимых статусов товара

    constraint uq_seller_products unique (seller_id, external_product_id), -- Уникальность товара у продавца
    constraint chk_seller_products_price check (price >= 0), -- Цена не может быть отрицательной
    constraint chk_seller_products_stock_qty check (stock_qty >= 0) -- Остаток не может быть отрицательным
);

comment on table seller_products is 'Товары продавцов';
comment on column seller_products.id is 'PK товара';
comment on column seller_products.seller_id is 'FK на продавца';
comment on column seller_products.external_product_id is 'ID товара во внешней системе';
comment on column seller_products.sku is 'Артикул / SKU';
comment on column seller_products.product_name is 'Название товара';
comment on column seller_products.category_id is 'FK на категорию';
comment on column seller_products.status is 'Статус товара';
comment on column seller_products.price is 'Цена товара';
comment on column seller_products.stock_qty is 'Остаток товара';
comment on column seller_products.published_at is 'Дата публикации';
comment on column seller_products.unpublished_at is 'Дата снятия с публикации';
comment on column seller_products.created_at is 'Дата создания записи';
comment on column seller_products.updated_at is 'Дата обновления записи';
comment on column seller_products.loaded_at is 'Дата загрузки из внешней системы';

-- ---------------------------------------------------------
-- Таблица заказов продавца
-- Загружается из внешней системы
-- ---------------------------------------------------------
create table seller_orders (
    id bigserial primary key, -- Внутренний идентификатор заказа
    seller_id bigint not null references sellers(id) on delete cascade, -- Ссылка на продавца
    external_order_id varchar(100) not null, -- Идентификатор заказа во внешней системе
    product_id bigint references seller_products(id), -- Ссылка на товар продавца
    region_id bigint references regions(id), -- Регион заказа
    order_status varchar(50) not null, -- Статус заказа
    order_amount numeric(14,2) not null default 0, -- Сумма заказа
    margin_amount numeric(14,2), -- Маржинальность заказа
    ordered_at timestamp with time zone not null, -- Дата оформления заказа
    completed_at timestamp with time zone, -- Дата завершения заказа
    cancelled_at timestamp with time zone, -- Дата отмены заказа
    created_at timestamp with time zone not null default now(), -- Дата создания записи
    updated_at timestamp with time zone not null default now(), -- Дата обновления записи
    loaded_at timestamp with time zone not null default now(), -- Дата загрузки из внешней системы

    constraint chk_seller_orders_status check (
        order_status in ('created', 'paid', 'completed', 'cancelled', 'returned')
    ), -- Ограничение допустимых статусов заказа

    constraint uq_seller_orders unique (seller_id, external_order_id), -- Уникальность заказа у продавца
    constraint chk_seller_orders_order_amount check (order_amount >= 0) -- Сумма заказа не может быть отрицательной
);

comment on table seller_orders is 'Заказы продавцов';
comment on column seller_orders.id is 'PK заказа';
comment on column seller_orders.seller_id is 'FK на продавца';
comment on column seller_orders.external_order_id is 'ID заказа во внешней системе';
comment on column seller_orders.product_id is 'FK на товар продавца';
comment on column seller_orders.region_id is 'FK на регион заказа';
comment on column seller_orders.order_status is 'Статус заказа';
comment on column seller_orders.order_amount is 'Сумма заказа';
comment on column seller_orders.margin_amount is 'Маржа по заказу';
comment on column seller_orders.ordered_at is 'Дата оформления заказа';
comment on column seller_orders.completed_at is 'Дата завершения заказа';
comment on column seller_orders.cancelled_at is 'Дата отмены заказа';
comment on column seller_orders.created_at is 'Дата создания записи';
comment on column seller_orders.updated_at is 'Дата обновления записи';
comment on column seller_orders.loaded_at is 'Дата загрузки из внешней системы';

-- ---------------------------------------------------------
-- Таблица активности продавца
-- Логи действий продавца в системе
-- ---------------------------------------------------------
create table seller_activity_log (
    id bigserial primary key, -- Внутренний идентификатор события активности
    seller_id bigint not null references sellers(id) on delete cascade, -- Ссылка на продавца
    activity_type varchar(100) not null, -- Тип активности продавца
    activity_at timestamp with time zone not null, -- Время действия
    source_system varchar(100), -- Источник события активности
    payload_json jsonb, -- Дополнительные данные события
    loaded_at timestamp with time zone not null default now() -- Дата загрузки события
);

comment on table seller_activity_log is 'Лог активности продавца';
comment on column seller_activity_log.id is 'PK события активности';
comment on column seller_activity_log.seller_id is 'FK на продавца';
comment on column seller_activity_log.activity_type is 'Тип действия';
comment on column seller_activity_log.activity_at is 'Время действия';
comment on column seller_activity_log.source_system is 'Источник события';
comment on column seller_activity_log.payload_json is 'Дополнительный payload';
comment on column seller_activity_log.loaded_at is 'Дата загрузки события';

-- ---------------------------------------------------------
-- Таблица присутствия в категориях
-- Упрощает анализ категорий продавца
-- ---------------------------------------------------------
create table seller_category_presence (
    id bigserial primary key, -- Внутренний идентификатор записи
    seller_id bigint not null references sellers(id) on delete cascade, -- Ссылка на продавца
    category_id bigint not null references categories(id), -- Ссылка на категорию
    first_seen_at timestamp with time zone not null, -- Дата первого появления продавца в категории
    last_seen_at timestamp with time zone not null, -- Дата последнего появления продавца в категории
    is_active boolean not null default true, -- Признак актуальности категории для продавца

    constraint uq_seller_category_presence unique (seller_id, category_id) -- Одна запись на пару продавец-категория
);

comment on table seller_category_presence is 'Наличие продавца в категориях';
comment on column seller_category_presence.id is 'PK записи';
comment on column seller_category_presence.seller_id is 'FK на продавца';
comment on column seller_category_presence.category_id is 'FK на категорию';
comment on column seller_category_presence.first_seen_at is 'Первое появление в категории';
comment on column seller_category_presence.last_seen_at is 'Последнее появление в категории';
comment on column seller_category_presence.is_active is 'Признак активности категории';

-- ---------------------------------------------------------
-- Таблица snapshot-метрик продавца
-- Текущее агрегированное состояние продавца
-- ---------------------------------------------------------
create table seller_metrics_snapshot (
    id bigserial primary key, -- Внутренний идентификатор snapshot
    seller_id bigint not null references sellers(id) on delete cascade, -- Ссылка на продавца
    snapshot_date date not null, -- Дата snapshot
    active_products_count integer not null default 0, -- Количество активных товаров
    published_products_count integer not null default 0, -- Количество опубликованных товаров
    products_without_stock_count integer not null default 0, -- Количество товаров без остатков
    categories_count integer not null default 0, -- Количество категорий продавца
    active_categories_count integer not null default 0, -- Количество активных категорий
    regions_count integer not null default 0, -- Количество регионов продаж
    orders_7d integer not null default 0, -- Количество заказов за 7 дней
    orders_30d integer not null default 0, -- Количество заказов за 30 дней
    revenue_7d numeric(14,2) not null default 0, -- Выручка за 7 дней
    revenue_30d numeric(14,2) not null default 0, -- Выручка за 30 дней
    margin_30d numeric(14,2) not null default 0, -- Маржа за 30 дней
    last_login_days integer not null default 0, -- Сколько дней прошло с последнего входа
    no_sales_days integer not null default 0, -- Сколько дней нет продаж
    current_primary_model_code varchar(50), -- Основная модель продаж продавца
    created_at timestamp with time zone not null default now(), -- Дата создания snapshot

    constraint uq_seller_metrics_snapshot unique (seller_id, snapshot_date), -- Один snapshot на продавца в день
    constraint chk_seller_metrics_snapshot_counts check (
        active_products_count >= 0 and
        published_products_count >= 0 and
        products_without_stock_count >= 0 and
        categories_count >= 0 and
        active_categories_count >= 0 and
        regions_count >= 0 and
        orders_7d >= 0 and
        orders_30d >= 0 and
        last_login_days >= 0 and
        no_sales_days >= 0
    ) -- Все счетчики должны быть неотрицательными
);

comment on table seller_metrics_snapshot is 'Агрегированный snapshot продавца';
comment on column seller_metrics_snapshot.id is 'PK snapshot';
comment on column seller_metrics_snapshot.seller_id is 'FK на продавца';
comment on column seller_metrics_snapshot.snapshot_date is 'Дата snapshot';
comment on column seller_metrics_snapshot.active_products_count is 'Активные товары';
comment on column seller_metrics_snapshot.published_products_count is 'Опубликованные товары';
comment on column seller_metrics_snapshot.products_without_stock_count is 'Товары без остатков';
comment on column seller_metrics_snapshot.categories_count is 'Количество категорий';
comment on column seller_metrics_snapshot.active_categories_count is 'Активные категории';
comment on column seller_metrics_snapshot.regions_count is 'Количество регионов';
comment on column seller_metrics_snapshot.orders_7d is 'Заказы за 7 дней';
comment on column seller_metrics_snapshot.orders_30d is 'Заказы за 30 дней';
comment on column seller_metrics_snapshot.revenue_7d is 'Выручка за 7 дней';
comment on column seller_metrics_snapshot.revenue_30d is 'Выручка за 30 дней';
comment on column seller_metrics_snapshot.margin_30d is 'Маржа за 30 дней';
comment on column seller_metrics_snapshot.last_login_days is 'Дней с последнего входа';
comment on column seller_metrics_snapshot.no_sales_days is 'Дней без продаж';
comment on column seller_metrics_snapshot.current_primary_model_code is 'Основная модель продаж';
comment on column seller_metrics_snapshot.created_at is 'Дата создания snapshot';

-- ---------------------------------------------------------
-- Таблица метрик по категориям
-- Агрегаты продавца по категориям
-- ---------------------------------------------------------
create table seller_category_metrics (
    id bigserial primary key, -- Внутренний идентификатор метрики
    seller_id bigint not null references sellers(id) on delete cascade, -- Ссылка на продавца
    category_id bigint not null references categories(id), -- Ссылка на категорию
    metric_date date not null, -- Дата метрики
    products_count integer not null default 0, -- Количество товаров в категории
    orders_count integer not null default 0, -- Количество заказов по категории
    revenue numeric(14,2) not null default 0, -- Выручка по категории
    margin numeric(14,2) not null default 0, -- Маржа по категории
    last_sale_at timestamp with time zone, -- Дата последней продажи в категории
    created_at timestamp with time zone not null default now(), -- Дата создания записи

    constraint uq_seller_category_metrics unique (seller_id, category_id, metric_date), -- Уникальная метрика на день
    constraint chk_seller_category_metrics_counts check (
        products_count >= 0 and
        orders_count >= 0
    ) -- Счетчики должны быть неотрицательными
);

comment on table seller_category_metrics is 'Метрики продавца по категориям';
comment on column seller_category_metrics.id is 'PK метрики';
comment on column seller_category_metrics.seller_id is 'FK на продавца';
comment on column seller_category_metrics.category_id is 'FK на категорию';
comment on column seller_category_metrics.metric_date is 'Дата метрики';
comment on column seller_category_metrics.products_count is 'Количество товаров';
comment on column seller_category_metrics.orders_count is 'Количество заказов';
comment on column seller_category_metrics.revenue is 'Выручка';
comment on column seller_category_metrics.margin is 'Маржа';
comment on column seller_category_metrics.last_sale_at is 'Последняя продажа';
comment on column seller_category_metrics.created_at is 'Дата создания записи';

-- ---------------------------------------------------------
-- Таблица метрик по регионам
-- Агрегаты продавца по регионам
-- ---------------------------------------------------------
create table seller_region_metrics (
    id bigserial primary key, -- Внутренний идентификатор метрики
    seller_id bigint not null references sellers(id) on delete cascade, -- Ссылка на продавца
    region_id bigint not null references regions(id), -- Ссылка на регион
    metric_date date not null, -- Дата метрики
    orders_count integer not null default 0, -- Количество заказов в регионе
    revenue numeric(14,2) not null default 0, -- Выручка в регионе
    margin numeric(14,2) not null default 0, -- Маржа в регионе
    avg_delivery_days numeric(8,2), -- Среднее время доставки в днях
    created_at timestamp with time zone not null default now(), -- Дата создания записи

    constraint uq_seller_region_metrics unique (seller_id, region_id, metric_date), -- Уникальная метрика на день
    constraint chk_seller_region_metrics_orders_count check (orders_count >= 0) -- Заказы не могут быть отрицательными
);

comment on table seller_region_metrics is 'Метрики продавца по регионам';
comment on column seller_region_metrics.id is 'PK метрики';
comment on column seller_region_metrics.seller_id is 'FK на продавца';
comment on column seller_region_metrics.region_id is 'FK на регион';
comment on column seller_region_metrics.metric_date is 'Дата метрики';
comment on column seller_region_metrics.orders_count is 'Количество заказов';
comment on column seller_region_metrics.revenue is 'Выручка';
comment on column seller_region_metrics.margin is 'Маржа';
comment on column seller_region_metrics.avg_delivery_days is 'Среднее время доставки';
comment on column seller_region_metrics.created_at is 'Дата создания записи';

-- ---------------------------------------------------------
-- Таблица feature-срезов
-- Задел под будущий AI/ML
-- ---------------------------------------------------------
create table seller_features (
    id bigserial primary key, -- Внутренний идентификатор feature-среза
    seller_id bigint not null references sellers(id) on delete cascade, -- Ссылка на продавца
    feature_date date not null, -- Дата feature-среза
    feature_version varchar(50) not null, -- Версия набора фичей
    features_json jsonb not null, -- Набор вычисленных фичей в JSON
    created_at timestamp with time zone not null default now(), -- Дата создания записи

    constraint uq_seller_features unique (seller_id, feature_date, feature_version) -- Уникальность версии фичей на дату
);

comment on table seller_features is 'Вычисленные фичи продавца';
comment on column seller_features.id is 'PK feature-среза';
comment on column seller_features.seller_id is 'FK на продавца';
comment on column seller_features.feature_date is 'Дата feature-среза';
comment on column seller_features.feature_version is 'Версия фичей';
comment on column seller_features.features_json is 'JSON с вычисленными фичами';
comment on column seller_features.created_at is 'Дата создания записи';

-- ---------------------------------------------------------
-- Таблица типов рекомендаций
-- Справочник бизнес-типов рекомендаций
-- ---------------------------------------------------------
create table recommendation_types (
    id bigserial primary key, -- Внутренний идентификатор типа рекомендации
    code varchar(100) not null unique, -- Уникальный код типа рекомендации
    name varchar(255) not null, -- Название типа рекомендации
    description text, -- Описание типа рекомендации
    is_active boolean not null default true -- Признак активности типа
);

comment on table recommendation_types is 'Справочник типов рекомендаций';
comment on column recommendation_types.id is 'PK типа рекомендации';
comment on column recommendation_types.code is 'Код типа рекомендации';
comment on column recommendation_types.name is 'Название типа рекомендации';
comment on column recommendation_types.description is 'Описание типа рекомендации';
comment on column recommendation_types.is_active is 'Признак активности типа';

-- ---------------------------------------------------------
-- Таблица правил рекомендаций
-- Бизнес-правила для генерации рекомендаций
-- ---------------------------------------------------------
create table recommendation_rules (
    id bigserial primary key, -- Внутренний идентификатор правила
    rule_code varchar(100) not null unique, -- Уникальный код правила
    rule_name varchar(255) not null, -- Название правила
    description text, -- Описание логики правила
    recommendation_type_id bigint not null references recommendation_types(id), -- Тип рекомендации, который создаёт правило
    priority integer not null default 100, -- Приоритет правила
    cooldown_days integer not null default 0, -- Период повторной блокировки в днях
    is_active boolean not null default true, -- Признак активности правила
    condition_expression text not null, -- Выражение условия правила
    created_by varchar(100), -- Кто создал правило
    updated_by varchar(100), -- Кто последний менял правило
    created_at timestamp with time zone not null default now(), -- Дата создания правила
    updated_at timestamp with time zone not null default now(), -- Дата обновления правила

    constraint chk_recommendation_rules_priority check (priority >= 0), -- Приоритет не может быть отрицательным
    constraint chk_recommendation_rules_cooldown_days check (cooldown_days >= 0) -- Cooldown не может быть отрицательным
);

comment on table recommendation_rules is 'Бизнес-правила рекомендаций';
comment on column recommendation_rules.id is 'PK правила';
comment on column recommendation_rules.rule_code is 'Код правила';
comment on column recommendation_rules.rule_name is 'Название правила';
comment on column recommendation_rules.description is 'Описание правила';
comment on column recommendation_rules.recommendation_type_id is 'FK на тип рекомендации';
comment on column recommendation_rules.priority is 'Приоритет правила';
comment on column recommendation_rules.cooldown_days is 'Cooldown в днях';
comment on column recommendation_rules.is_active is 'Признак активности правила';
comment on column recommendation_rules.condition_expression is 'Условие правила';
comment on column recommendation_rules.created_by is 'Автор создания';
comment on column recommendation_rules.updated_by is 'Автор последнего изменения';
comment on column recommendation_rules.created_at is 'Дата создания правила';
comment on column recommendation_rules.updated_at is 'Дата обновления правила';

-- ---------------------------------------------------------
-- Таблица шаблонов уведомлений
-- Текстовые шаблоны коммуникации
-- ---------------------------------------------------------
create table notification_templates (
    id bigserial primary key, -- Внутренний идентификатор шаблона
    template_code varchar(100) not null unique, -- Уникальный код шаблона
    recommendation_type_id bigint not null references recommendation_types(id), -- Тип рекомендации для шаблона
    title_template varchar(500) not null, -- Шаблон заголовка уведомления
    body_template text not null, -- Шаблон текста уведомления
    channel_code varchar(50) not null, -- Канал доставки шаблона
    is_active boolean not null default true, -- Признак активности шаблона
    created_at timestamp with time zone not null default now(), -- Дата создания шаблона
    updated_at timestamp with time zone not null default now() -- Дата обновления шаблона
);

comment on table notification_templates is 'Шаблоны уведомлений';
comment on column notification_templates.id is 'PK шаблона';
comment on column notification_templates.template_code is 'Код шаблона';
comment on column notification_templates.recommendation_type_id is 'FK на тип рекомендации';
comment on column notification_templates.title_template is 'Шаблон заголовка';
comment on column notification_templates.body_template is 'Шаблон тела уведомления';
comment on column notification_templates.channel_code is 'Канал доставки';
comment on column notification_templates.is_active is 'Признак активности шаблона';
comment on column notification_templates.created_at is 'Дата создания шаблона';
comment on column notification_templates.updated_at is 'Дата обновления шаблона';

-- ---------------------------------------------------------
-- Таблица логов срабатывания триггеров
-- Факт срабатывания правила для продавца
-- ---------------------------------------------------------
create table seller_trigger_log (
    id bigserial primary key, -- Внутренний идентификатор триггера
    seller_id bigint not null references sellers(id) on delete cascade, -- Ссылка на продавца
    rule_id bigint not null references recommendation_rules(id), -- Ссылка на правило
    trigger_code varchar(100) not null, -- Код сработавшего триггера
    triggered_at timestamp with time zone not null default now(), -- Время срабатывания триггера
    period_key varchar(50) not null, -- Период дедупликации: день/неделя/месяц
    snapshot_id bigint references seller_metrics_snapshot(id), -- Snapshot, на основе которого сработал триггер
    payload_json jsonb, -- Контекст срабатывания триггера
    status varchar(50) not null default 'detected', -- Статус обработки триггера
    created_at timestamp with time zone not null default now(), -- Дата создания записи

    constraint chk_seller_trigger_log_status check (
        status in ('detected', 'skipped', 'converted_to_recommendation')
    ), -- Допустимые статусы триггера

    constraint uq_seller_trigger_log unique (seller_id, rule_id, period_key) -- Один триггер на правило и период
);

comment on table seller_trigger_log is 'Лог срабатывания триггеров';
comment on column seller_trigger_log.id is 'PK триггера';
comment on column seller_trigger_log.seller_id is 'FK на продавца';
comment on column seller_trigger_log.rule_id is 'FK на правило';
comment on column seller_trigger_log.trigger_code is 'Код триггера';
comment on column seller_trigger_log.triggered_at is 'Время срабатывания';
comment on column seller_trigger_log.period_key is 'Ключ периода дедупликации';
comment on column seller_trigger_log.snapshot_id is 'FK на snapshot продавца';
comment on column seller_trigger_log.payload_json is 'Контекст триггера';
comment on column seller_trigger_log.status is 'Статус обработки триггера';
comment on column seller_trigger_log.created_at is 'Дата создания записи';

-- ---------------------------------------------------------
-- Таблица рекомендаций
-- Бизнес-рекомендации, созданные системой
-- ---------------------------------------------------------
create table recommendations (
    id bigserial primary key, -- Внутренний идентификатор рекомендации
    seller_id bigint not null references sellers(id) on delete cascade, -- Ссылка на продавца
    trigger_id bigint not null references seller_trigger_log(id) on delete cascade, -- Ссылка на триггер
    recommendation_type_id bigint not null references recommendation_types(id), -- Тип рекомендации
    template_id bigint references notification_templates(id), -- Использованный шаблон уведомления
    title varchar(500) not null, -- Заголовок рекомендации
    description text not null, -- Описание рекомендации
    reason_text text, -- Объяснение, почему рекомендация была создана
    priority integer not null default 100, -- Приоритет рекомендации
    score numeric(8,4), -- Скоринг рекомендации
    status varchar(50) not null default 'created', -- Статус рекомендации
    expires_at timestamp with time zone, -- Время истечения актуальности рекомендации
    created_at timestamp with time zone not null default now(), -- Дата создания рекомендации
    updated_at timestamp with time zone not null default now(), -- Дата обновления рекомендации

    constraint chk_recommendations_status check (
        status in ('created', 'ready_to_send', 'sent', 'opened', 'accepted', 'rejected', 'expired')
    ), -- Допустимые статусы рекомендации

    constraint chk_recommendations_priority check (priority >= 0) -- Приоритет не может быть отрицательным
);

comment on table recommendations is 'Сформированные рекомендации';
comment on column recommendations.id is 'PK рекомендации';
comment on column recommendations.seller_id is 'FK на продавца';
comment on column recommendations.trigger_id is 'FK на триггер';
comment on column recommendations.recommendation_type_id is 'FK на тип рекомендации';
comment on column recommendations.template_id is 'FK на шаблон уведомления';
comment on column recommendations.title is 'Заголовок рекомендации';
comment on column recommendations.description is 'Описание рекомендации';
comment on column recommendations.reason_text is 'Причина создания рекомендации';
comment on column recommendations.priority is 'Приоритет рекомендации';
comment on column recommendations.score is 'Скоринговая оценка';
comment on column recommendations.status is 'Статус рекомендации';
comment on column recommendations.expires_at is 'Время истечения рекомендации';
comment on column recommendations.created_at is 'Дата создания рекомендации';
comment on column recommendations.updated_at is 'Дата обновления рекомендации';

-- ---------------------------------------------------------
-- Таблица логов уведомлений
-- Факт доставки/отправки уведомления по рекомендации
-- ---------------------------------------------------------
create table seller_notification_log (
    id bigserial primary key, -- Внутренний идентификатор уведомления
    seller_id bigint not null references sellers(id) on delete cascade, -- Ссылка на продавца
    recommendation_id bigint not null references recommendations(id) on delete cascade, -- Ссылка на рекомендацию
    channel_code varchar(50) not null, -- Канал уведомления
    delivery_system_id varchar(100), -- Идентификатор уведомления во внешнем сервисе доставки
    status varchar(50) not null default 'created', -- Статус уведомления
    payload_json jsonb, -- Payload, отправленный в сервис доставки
    sent_at timestamp with time zone, -- Время отправки
    delivered_at timestamp with time zone, -- Время доставки
    opened_at timestamp with time zone, -- Время открытия
    clicked_at timestamp with time zone, -- Время клика
    error_message text, -- Ошибка доставки, если была
    created_at timestamp with time zone not null default now(), -- Дата создания уведомления
    updated_at timestamp with time zone not null default now(), -- Дата обновления уведомления

    constraint chk_seller_notification_log_status check (
        status in ('created', 'ready_to_send', 'sent', 'delivered', 'failed', 'opened', 'clicked')
    ) -- Допустимые статусы уведомления
);

comment on table seller_notification_log is 'Лог уведомлений';
comment on column seller_notification_log.id is 'PK уведомления';
comment on column seller_notification_log.seller_id is 'FK на продавца';
comment on column seller_notification_log.recommendation_id is 'FK на рекомендацию';
comment on column seller_notification_log.channel_code is 'Канал доставки';
comment on column seller_notification_log.delivery_system_id is 'ID во внешнем сервисе доставки';
comment on column seller_notification_log.status is 'Статус уведомления';
comment on column seller_notification_log.payload_json is 'Payload уведомления';
comment on column seller_notification_log.sent_at is 'Время отправки';
comment on column seller_notification_log.delivered_at is 'Время доставки';
comment on column seller_notification_log.opened_at is 'Время открытия';
comment on column seller_notification_log.clicked_at is 'Время клика';
comment on column seller_notification_log.error_message is 'Описание ошибки';
comment on column seller_notification_log.created_at is 'Дата создания уведомления';
comment on column seller_notification_log.updated_at is 'Дата обновления уведомления';

-- ---------------------------------------------------------
-- Таблица обратной связи по рекомендациям
-- История действий пользователя
-- ---------------------------------------------------------
create table recommendation_feedback (
    id bigserial primary key, -- Внутренний идентификатор feedback
    seller_id bigint not null references sellers(id) on delete cascade, -- Ссылка на продавца
    recommendation_id bigint not null references recommendations(id) on delete cascade, -- Ссылка на рекомендацию
    feedback_type varchar(50) not null, -- Тип действия пользователя
    feedback_at timestamp with time zone not null default now(), -- Время действия пользователя
    payload_json jsonb, -- Дополнительный контекст действия
    created_at timestamp with time zone not null default now(), -- Дата создания feedback

    constraint chk_recommendation_feedback_type check (
        feedback_type in ('view', 'accept', 'reject', 'click', 'dismiss')
    ) -- Допустимые типы feedback
);

comment on table recommendation_feedback is 'Обратная связь по рекомендациям';
comment on column recommendation_feedback.id is 'PK feedback';
comment on column recommendation_feedback.seller_id is 'FK на продавца';
comment on column recommendation_feedback.recommendation_id is 'FK на рекомендацию';
comment on column recommendation_feedback.feedback_type is 'Тип действия';
comment on column recommendation_feedback.feedback_at is 'Время действия';
comment on column recommendation_feedback.payload_json is 'Дополнительный payload';
comment on column recommendation_feedback.created_at is 'Дата создания feedback';

-- ---------------------------------------------------------
-- Таблица загрузок данных
-- История job-ов Data Loader
-- ---------------------------------------------------------
create table data_load_jobs (
    id bigserial primary key, -- Внутренний идентификатор job загрузки
    job_type varchar(100) not null, -- Тип job загрузки
    status varchar(50) not null, -- Статус job
    started_at timestamp with time zone, -- Время начала выполнения
    finished_at timestamp with time zone, -- Время завершения выполнения
    records_loaded integer not null default 0, -- Количество успешно загруженных записей
    records_failed integer not null default 0, -- Количество ошибочных записей
    error_message text, -- Текст ошибки job
    created_at timestamp with time zone not null default now(), -- Дата создания job

    constraint chk_data_load_jobs_status check (
        status in ('created', 'running', 'success', 'failed', 'partial_success')
    ), -- Допустимые статусы job загрузки

    constraint chk_data_load_jobs_records check (
        records_loaded >= 0 and records_failed >= 0
    ) -- Счетчики job не могут быть отрицательными
);

comment on table data_load_jobs is 'История job загрузки данных';
comment on column data_load_jobs.id is 'PK job загрузки';
comment on column data_load_jobs.job_type is 'Тип job';
comment on column data_load_jobs.status is 'Статус job';
comment on column data_load_jobs.started_at is 'Время начала';
comment on column data_load_jobs.finished_at is 'Время завершения';
comment on column data_load_jobs.records_loaded is 'Успешно загружено записей';
comment on column data_load_jobs.records_failed is 'Ошибочных записей';
comment on column data_load_jobs.error_message is 'Описание ошибки';
comment on column data_load_jobs.created_at is 'Дата создания job';

-- ---------------------------------------------------------
-- Таблица job-ов анализа
-- История batch-анализа продавцов
-- ---------------------------------------------------------
create table analysis_jobs (
    id bigserial primary key, -- Внутренний идентификатор job анализа
    job_type varchar(100) not null, -- Тип job анализа
    status varchar(50) not null, -- Статус job анализа
    started_at timestamp with time zone, -- Время начала анализа
    finished_at timestamp with time zone, -- Время завершения анализа
    sellers_processed integer not null default 0, -- Количество обработанных продавцов
    recommendations_created integer not null default 0, -- Количество созданных рекомендаций
    triggers_created integer not null default 0, -- Количество созданных триггеров
    error_message text, -- Описание ошибки job
    created_at timestamp with time zone not null default now(), -- Дата создания job

    constraint chk_analysis_jobs_status check (
        status in ('created', 'running', 'success', 'failed', 'partial_success')
    ), -- Допустимые статусы job анализа

    constraint chk_analysis_jobs_counts check (
        sellers_processed >= 0 and
        recommendations_created >= 0 and
        triggers_created >= 0
    ) -- Счетчики анализа не могут быть отрицательными
);

comment on table analysis_jobs is 'История job анализа';
comment on column analysis_jobs.id is 'PK job анализа';
comment on column analysis_jobs.job_type is 'Тип job';
comment on column analysis_jobs.status is 'Статус job';
comment on column analysis_jobs.started_at is 'Время начала';
comment on column analysis_jobs.finished_at is 'Время завершения';
comment on column analysis_jobs.sellers_processed is 'Обработано продавцов';
comment on column analysis_jobs.recommendations_created is 'Создано рекомендаций';
comment on column analysis_jobs.triggers_created is 'Создано триггеров';
comment on column analysis_jobs.error_message is 'Описание ошибки';
comment on column analysis_jobs.created_at is 'Дата создания job';

-- ---------------------------------------------------------
-- Таблица outbox-очереди уведомлений
-- Надежная отправка в сервис доставки уведомлений
-- ---------------------------------------------------------
create table outbound_notification_queue (
    id bigserial primary key, -- Внутренний идентификатор очереди
    recommendation_id bigint not null references recommendations(id) on delete cascade, -- Ссылка на рекомендацию
    notification_id bigint not null references seller_notification_log(id) on delete cascade, -- Ссылка на лог уведомления
    status varchar(50) not null default 'pending', -- Статус записи в очереди
    attempt_count integer not null default 0, -- Количество попыток отправки
    next_retry_at timestamp with time zone, -- Время следующей повторной попытки
    last_error text, -- Последняя ошибка отправки
    created_at timestamp with time zone not null default now(), -- Дата создания записи
    updated_at timestamp with time zone not null default now(), -- Дата обновления записи

    constraint chk_outbound_notification_queue_status check (
        status in ('pending', 'sent', 'failed')
    ), -- Допустимые статусы очереди

    constraint chk_outbound_notification_queue_attempt_count check (attempt_count >= 0), -- Количество попыток не может быть отрицательным
    constraint uq_outbound_notification_queue_notification_id unique (notification_id) -- Одна outbox-запись на одно уведомление
);

comment on table outbound_notification_queue is 'Outbox-очередь уведомлений';
comment on column outbound_notification_queue.id is 'PK очереди';
comment on column outbound_notification_queue.recommendation_id is 'FK на рекомендацию';
comment on column outbound_notification_queue.notification_id is 'FK на уведомление';
comment on column outbound_notification_queue.status is 'Статус outbox-записи';
comment on column outbound_notification_queue.attempt_count is 'Количество попыток отправки';
comment on column outbound_notification_queue.next_retry_at is 'Следующее время retry';
comment on column outbound_notification_queue.last_error is 'Последняя ошибка';
comment on column outbound_notification_queue.created_at is 'Дата создания записи';
comment on column outbound_notification_queue.updated_at is 'Дата обновления записи';

-- =========================================================
-- Индексы для производительности
-- =========================================================

create index idx_sellers_external_seller_id on sellers(external_seller_id); -- Быстрый поиск продавца по внешнему ID
create index idx_sellers_home_region_id on sellers(home_region_id); -- Поиск продавцов по региону
create index idx_sellers_lifecycle_stage on sellers(lifecycle_stage); -- Поиск по стадии жизненного цикла

create index idx_seller_models_seller_id on seller_models(seller_id); -- Поиск моделей продавца
create index idx_seller_models_model_code on seller_models(model_code); -- Поиск по коду модели

create index idx_seller_products_seller_id on seller_products(seller_id); -- Поиск товаров продавца
create index idx_seller_products_category_id on seller_products(category_id); -- Поиск товаров по категории
create index idx_seller_products_status on seller_products(status); -- Поиск товаров по статусу
create index idx_seller_products_loaded_at on seller_products(loaded_at); -- Поиск по времени загрузки

create index idx_seller_orders_seller_id on seller_orders(seller_id); -- Поиск заказов продавца
create index idx_seller_orders_product_id on seller_orders(product_id); -- Поиск заказов по товару
create index idx_seller_orders_region_id on seller_orders(region_id); -- Поиск заказов по региону
create index idx_seller_orders_order_status on seller_orders(order_status); -- Поиск заказов по статусу
create index idx_seller_orders_ordered_at on seller_orders(ordered_at); -- Поиск по дате заказа
create index idx_seller_orders_loaded_at on seller_orders(loaded_at); -- Поиск по времени загрузки

create index idx_seller_activity_log_seller_id on seller_activity_log(seller_id); -- Поиск активности продавца
create index idx_seller_activity_log_activity_type on seller_activity_log(activity_type); -- Поиск по типу активности
create index idx_seller_activity_log_activity_at on seller_activity_log(activity_at); -- Поиск по времени активности

create index idx_seller_category_presence_seller_id on seller_category_presence(seller_id); -- Поиск категорий продавца
create index idx_seller_category_presence_category_id on seller_category_presence(category_id); -- Поиск продавцов по категории

create index idx_seller_metrics_snapshot_seller_id on seller_metrics_snapshot(seller_id); -- Поиск snapshot продавца
create index idx_seller_metrics_snapshot_snapshot_date on seller_metrics_snapshot(snapshot_date); -- Поиск snapshot по дате

create index idx_seller_category_metrics_seller_id on seller_category_metrics(seller_id); -- Поиск категорийных метрик продавца
create index idx_seller_category_metrics_category_id on seller_category_metrics(category_id); -- Поиск метрик по категории
create index idx_seller_category_metrics_metric_date on seller_category_metrics(metric_date); -- Поиск метрик по дате

create index idx_seller_region_metrics_seller_id on seller_region_metrics(seller_id); -- Поиск региональных метрик продавца
create index idx_seller_region_metrics_region_id on seller_region_metrics(region_id); -- Поиск метрик по региону
create index idx_seller_region_metrics_metric_date on seller_region_metrics(metric_date); -- Поиск метрик по дате

create index idx_seller_features_seller_id on seller_features(seller_id); -- Поиск feature-срезов продавца
create index idx_seller_features_feature_date on seller_features(feature_date); -- Поиск feature-срезов по дате

create index idx_recommendation_rules_is_active on recommendation_rules(is_active); -- Поиск активных правил
create index idx_recommendation_rules_recommendation_type_id on recommendation_rules(recommendation_type_id); -- Поиск правил по типу рекомендации

create index idx_notification_templates_recommendation_type_id on notification_templates(recommendation_type_id); -- Поиск шаблонов по типу рекомендации
create index idx_notification_templates_channel_code on notification_templates(channel_code); -- Поиск шаблонов по каналу

create index idx_seller_trigger_log_seller_id on seller_trigger_log(seller_id); -- Поиск триггеров продавца
create index idx_seller_trigger_log_rule_id on seller_trigger_log(rule_id); -- Поиск по правилу
create index idx_seller_trigger_log_triggered_at on seller_trigger_log(triggered_at); -- Поиск по времени триггера
create index idx_seller_trigger_log_status on seller_trigger_log(status); -- Поиск по статусу триггера

create index idx_recommendations_seller_id on recommendations(seller_id); -- Поиск рекомендаций продавца
create index idx_recommendations_trigger_id on recommendations(trigger_id); -- Поиск рекомендаций по триггеру
create index idx_recommendations_recommendation_type_id on recommendations(recommendation_type_id); -- Поиск по типу рекомендации
create index idx_recommendations_status on recommendations(status); -- Поиск по статусу рекомендации
create index idx_recommendations_created_at on recommendations(created_at); -- Поиск по дате создания

create index idx_seller_notification_log_seller_id on seller_notification_log(seller_id); -- Поиск уведомлений продавца
create index idx_seller_notification_log_recommendation_id on seller_notification_log(recommendation_id); -- Поиск уведомлений по рекомендации
create index idx_seller_notification_log_status on seller_notification_log(status); -- Поиск по статусу уведомления
create index idx_seller_notification_log_sent_at on seller_notification_log(sent_at); -- Поиск по времени отправки

create index idx_recommendation_feedback_seller_id on recommendation_feedback(seller_id); -- Поиск feedback продавца
create index idx_recommendation_feedback_recommendation_id on recommendation_feedback(recommendation_id); -- Поиск feedback по рекомендации
create index idx_recommendation_feedback_feedback_type on recommendation_feedback(feedback_type); -- Поиск по типу feedback
create index idx_recommendation_feedback_feedback_at on recommendation_feedback(feedback_at); -- Поиск по времени feedback

create index idx_data_load_jobs_status on data_load_jobs(status); -- Поиск job загрузки по статусу
create index idx_data_load_jobs_created_at on data_load_jobs(created_at); -- Поиск job загрузки по дате создания

create index idx_analysis_jobs_status on analysis_jobs(status); -- Поиск job анализа по статусу
create index idx_analysis_jobs_created_at on analysis_jobs(created_at); -- Поиск job анализа по дате создания

create index idx_outbound_notification_queue_status on outbound_notification_queue(status); -- Поиск outbox по статусу
create index idx_outbound_notification_queue_next_retry_at on outbound_notification_queue(next_retry_at); -- Поиск записей для retry
create index idx_outbound_notification_queue_created_at on outbound_notification_queue(created_at); -- Поиск по дате создания