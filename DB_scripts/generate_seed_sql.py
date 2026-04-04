#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
Генератор тестовых данных для Seller Recommendation Platform.

Результат:
    seed_data.sql

Как работает:
    1. Генерирует справочники
    2. Генерирует продавцов и связанные сущности
    3. Пишет SQL INSERT-скрипт

Настройка объема данных:
    Меняйте константы в блоке CONFIG
"""

from __future__ import annotations

import json
import random
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Iterable

# =========================================================
# CONFIG
# =========================================================

OUTPUT_FILE = "seed_data.sql"
RANDOM_SEED = 42

SELLERS_COUNT = 200
REGIONS_COUNT = 20
CATEGORIES_COUNT = 30

PRODUCTS_PER_SELLER_MIN = 1
PRODUCTS_PER_SELLER_MAX = 10

ORDERS_PER_SELLER_MIN = 0
ORDERS_PER_SELLER_MAX = 20

ACTIVITY_PER_SELLER_MIN = 2
ACTIVITY_PER_SELLER_MAX = 12

SELLER_MODELS_PER_SELLER_MIN = 1
SELLER_MODELS_PER_SELLER_MAX = 3

CATEGORY_METRICS_PER_SELLER_MIN = 1
CATEGORY_METRICS_PER_SELLER_MAX = 5

REGION_METRICS_PER_SELLER_MIN = 1
REGION_METRICS_PER_SELLER_MAX = 5

RECOMMENDATIONS_PER_SELLER_MIN = 0
RECOMMENDATIONS_PER_SELLER_MAX = 3

FEATURES_PER_SELLER = 1
SNAPSHOTS_DAYS = 3

MAX_FEEDBACK_PER_RECOMMENDATION = 2
MAX_NOTIFICATIONS_PER_RECOMMENDATION = 1

MAX_OUTBOX_RETRIES = 3

# =========================================================
# RANDOM HELPERS
# =========================================================

random.seed(RANDOM_SEED)

UTC = timezone.utc
NOW = datetime.now(tz=UTC)


def rand_dt(days_back: int = 180) -> datetime:
    """Случайная дата в прошлом."""
    delta_days = random.randint(0, days_back)
    delta_seconds = random.randint(0, 24 * 60 * 60 - 1)
    return NOW - timedelta(days=delta_days, seconds=delta_seconds)


def rand_date(days_back: int = 180):
    """Случайная дата без времени."""
    return rand_dt(days_back).date()


def chance(probability: float) -> bool:
    """Вероятность true."""
    return random.random() < probability


def pick(seq):
    """Случайный элемент."""
    return random.choice(seq)


def sql_escape(value: str) -> str:
    """Экранирование строки для SQL."""
    return value.replace("'", "''")


def sql_str(value: str | None) -> str:
    """Преобразование Python string -> SQL."""
    if value is None:
        return "NULL"
    return f"'{sql_escape(value)}'"


def sql_bool(value: bool) -> str:
    """Преобразование bool -> SQL."""
    return "true" if value else "false"


def sql_num(value) -> str:
    """Преобразование числа -> SQL."""
    if value is None:
        return "NULL"
    return str(value)


def sql_ts(value: datetime | None) -> str:
    """Преобразование datetime -> SQL timestamp."""
    if value is None:
        return "NULL"
    return f"'{value.isoformat()}'"


def sql_date(value) -> str:
    """Преобразование date -> SQL date."""
    if value is None:
        return "NULL"
    return f"'{value.isoformat()}'"


def sql_json(value) -> str:
    """Преобразование объекта -> SQL jsonb."""
    if value is None:
        return "NULL"
    data = json.dumps(value, ensure_ascii=False)
    return f"'{sql_escape(data)}'::jsonb"


def money(min_v: float = 100.0, max_v: float = 10000.0) -> float:
    """Случайная денежная сумма."""
    return round(random.uniform(min_v, max_v), 2)


# =========================================================
# SQL WRITER
# =========================================================

class SqlWriter:
    """Простой writer для SQL-файла."""

    def __init__(self) -> None:
        self.lines: list[str] = []

    def line(self, text: str = "") -> None:
        self.lines.append(text)

    def comment(self, text: str) -> None:
        self.lines.append(f"-- {text}")

    def insert(self, table: str, columns: list[str], values: list[str]) -> None:
        cols = ", ".join(columns)
        vals = ", ".join(values)
        self.lines.append(f"insert into {table} ({cols}) values ({vals});")

    def save(self, path: str | Path) -> None:
        Path(path).write_text("\n".join(self.lines) + "\n", encoding="utf-8")


# =========================================================
# ID REGISTRIES
# =========================================================

@dataclass
class Registry:
    regions: list[int]
    categories: list[int]
    sellers: list[int]
    seller_products: list[int]
    seller_orders: list[int]
    seller_snapshots: list[int]
    recommendation_types: list[int]
    rules: list[int]
    templates: list[int]
    triggers: list[int]
    recommendations: list[int]
    notifications: list[int]

    def __init__(self) -> None:
        self.regions = []
        self.categories = []
        self.sellers = []
        self.seller_products = []
        self.seller_orders = []
        self.seller_snapshots = []
        self.recommendation_types = []
        self.rules = []
        self.templates = []
        self.triggers = []
        self.recommendations = []
        self.notifications = []


# =========================================================
# STATIC DATA
# =========================================================

SELLER_TYPES = ["company", "ip", "self_employed"]
SELLER_STATUSES = ["active", "inactive", "pending"]
LIFECYCLE_STAGES = [
    "registered",
    "catalog_setup",
    "first_sales",
    "active",
    "growth",
    "stagnation",
    "churn_risk",
    "reactivated",
]
MODEL_CODES = ["CLICK_COLLECT", "DBS", "FBS", "EXPRESS"]
MODEL_STATUSES = ["enabled", "disabled", "recommended", "unavailable"]

PRODUCT_STATUSES = ["draft", "active", "inactive", "archived"]
ORDER_STATUSES = ["created", "paid", "completed", "cancelled", "returned"]
ACTIVITY_TYPES = [
    "login",
    "open_dashboard",
    "edit_product",
    "publish_product",
    "view_analytics",
    "change_price",
]

RECOMMENDATION_TYPE_DATA = [
    ("ADD_PRODUCTS", "Добавить товары", "Рекомендация добавить новые товары"),
    ("EXPAND_GEO", "Расширить географию", "Рекомендация расширить регионы продаж"),
    ("SWITCH_MODEL", "Сменить модель продаж", "Рекомендация подключить другую модель"),
    ("ADD_CATEGORY", "Добавить категорию", "Рекомендация добавить новую категорию"),
    ("RESTOCK_PRODUCTS", "Пополнить остатки", "Рекомендация пополнить остатки"),
    ("REACTIVATE_ACTIVITY", "Вернуться к активности", "Рекомендация вернуться к работе"),
]

RULE_DATA = [
    ("NO_PRODUCTS_7D", "Нет товаров 7 дней", "Если продавец зарегистрирован, но нет товаров", "ADD_PRODUCTS"),
    ("NO_SALES_14D", "Нет продаж 14 дней", "Если есть товары, но нет продаж", "SWITCH_MODEL"),
    ("INACTIVE_7D", "Нет активности 7 дней", "Если продавец давно не заходил", "REACTIVATE_ACTIVITY"),
    ("HIGH_DEMAND_REGION", "Высокий спрос в других регионах", "Если стоит расширить географию", "EXPAND_GEO"),
    ("NEW_CATEGORY_OPPORTUNITY", "Новая категория", "Если стоит расширить категории", "ADD_CATEGORY"),
    ("OUT_OF_STOCK", "Нет остатков", "Если много товаров без остатков", "RESTOCK_PRODUCTS"),
]

CHANNEL_CODES = ["in_app", "push", "email"]

TRIGGER_CODES = [
    "NO_PRODUCTS",
    "NO_SALES",
    "INACTIVE_SELLER",
    "HIGH_DEMAND_REGION",
    "NEW_CATEGORY",
    "OUT_OF_STOCK",
]

RECOMMENDATION_STATUSES = [
    "created",
    "ready_to_send",
    "sent",
    "opened",
    "accepted",
    "rejected",
    "expired",
]

NOTIFICATION_STATUSES = [
    "created",
    "ready_to_send",
    "sent",
    "delivered",
    "failed",
    "opened",
    "clicked",
]

FEEDBACK_TYPES = ["view", "accept", "reject", "click", "dismiss"]
JOB_STATUSES = ["created", "running", "success", "failed", "partial_success"]
OUTBOX_STATUSES = ["pending", "sent", "failed"]


# =========================================================
# GENERATORS
# =========================================================

def write_header(w: SqlWriter) -> None:
    w.comment("=========================================================")
    w.comment("AUTO-GENERATED SEED DATA")
    w.comment("Seller Recommendation Platform")
    w.comment("=========================================================")
    w.line("begin;")
    w.line()


def write_cleanup(w: SqlWriter) -> None:
    w.comment("Очистка таблиц в обратном порядке зависимостей")
    tables = [
        "outbound_notification_queue",
        "recommendation_feedback",
        "seller_notification_log",
        "recommendations",
        "seller_trigger_log",
        "notification_templates",
        "recommendation_rules",
        "recommendation_types",
        "seller_features",
        "seller_region_metrics",
        "seller_category_metrics",
        "seller_metrics_snapshot",
        "seller_category_presence",
        "seller_activity_log",
        "seller_orders",
        "seller_products",
        "seller_models",
        "analysis_jobs",
        "data_load_jobs",
        "sellers",
        "categories",
        "regions",
    ]
    for table in tables:
        w.line(f"delete from {table};")
    w.line()


def write_regions(w: SqlWriter, reg: Registry) -> None:
    w.comment("Справочник регионов")
    for i in range(1, REGIONS_COUNT + 1):
        reg.regions.append(i)
        w.insert(
            "regions",
            ["id", "code", "name", "is_active"],
            [
                sql_num(i),
                sql_str(f"REGION_{i:03d}"),
                sql_str(f"Регион {i}"),
                sql_bool(True),
            ],
        )
    w.line()


def write_categories(w: SqlWriter, reg: Registry) -> None:
    w.comment("Справочник категорий")
    for i in range(1, CATEGORIES_COUNT + 1):
        reg.categories.append(i)
        parent_id = None
        if i > 5 and chance(0.35):
            parent_id = random.randint(1, i - 1)

        w.insert(
            "categories",
            ["id", "external_category_id", "name", "parent_id", "is_active"],
            [
                sql_num(i),
                sql_str(f"EXT_CAT_{i:03d}"),
                sql_str(f"Категория {i}"),
                sql_num(parent_id),
                sql_bool(True),
            ],
        )
    w.line()


def write_recommendation_types(w: SqlWriter, reg: Registry) -> dict[str, int]:
    w.comment("Справочник типов рекомендаций")
    code_to_id: dict[str, int] = {}
    for i, (code, name, description) in enumerate(RECOMMENDATION_TYPE_DATA, start=1):
        reg.recommendation_types.append(i)
        code_to_id[code] = i
        w.insert(
            "recommendation_types",
            ["id", "code", "name", "description", "is_active"],
            [
                sql_num(i),
                sql_str(code),
                sql_str(name),
                sql_str(description),
                sql_bool(True),
            ],
        )
    w.line()
    return code_to_id


def write_rules(w: SqlWriter, reg: Registry, rec_type_map: dict[str, int]) -> None:
    w.comment("Бизнес-правила рекомендаций")
    for i, (rule_code, rule_name, description, rec_type_code) in enumerate(RULE_DATA, start=1):
        reg.rules.append(i)
        w.insert(
            "recommendation_rules",
            [
                "id",
                "rule_code",
                "rule_name",
                "description",
                "recommendation_type_id",
                "priority",
                "cooldown_days",
                "is_active",
                "condition_expression",
                "created_by",
                "updated_by",
                "created_at",
                "updated_at",
            ],
            [
                sql_num(i),
                sql_str(rule_code),
                sql_str(rule_name),
                sql_str(description),
                sql_num(rec_type_map[rec_type_code]),
                sql_num(random.randint(50, 300)),
                sql_num(random.choice([3, 7, 14, 30])),
                sql_bool(True),
                sql_str(f"seller_metric_check::{rule_code}"),
                sql_str("seed_generator"),
                sql_str("seed_generator"),
                sql_ts(rand_dt(60)),
                sql_ts(rand_dt(30)),
            ],
        )
    w.line()


def write_templates(w: SqlWriter, reg: Registry, rec_type_map: dict[str, int]) -> None:
    w.comment("Шаблоны уведомлений")
    template_id = 1
    for code, _, _ in RECOMMENDATION_TYPE_DATA:
        for channel in CHANNEL_CODES:
            reg.templates.append(template_id)
            w.insert(
                "notification_templates",
                [
                    "id",
                    "template_code",
                    "recommendation_type_id",
                    "title_template",
                    "body_template",
                    "channel_code",
                    "is_active",
                    "created_at",
                    "updated_at",
                ],
                [
                    sql_num(template_id),
                    sql_str(f"{code}_{channel}".upper()),
                    sql_num(rec_type_map[code]),
                    sql_str(f"Рекомендация: {code}"),
                    sql_str(f"Система рекомендует выполнить действие {code} через канал {channel}"),
                    sql_str(channel),
                    sql_bool(True),
                    sql_ts(rand_dt(60)),
                    sql_ts(rand_dt(10)),
                ],
            )
            template_id += 1
    w.line()


def write_jobs(w: SqlWriter) -> None:
    w.comment("История job-ов загрузки данных")
    for i in range(1, 6):
        started = rand_dt(30)
        finished = started + timedelta(minutes=random.randint(1, 20))
        status = pick(JOB_STATUSES)
        w.insert(
            "data_load_jobs",
            [
                "id",
                "job_type",
                "status",
                "started_at",
                "finished_at",
                "records_loaded",
                "records_failed",
                "error_message",
                "created_at",
            ],
            [
                sql_num(i),
                sql_str(pick(["full_sync", "delta_sync", "seller_sync", "orders_sync"])),
                sql_str(status),
                sql_ts(started),
                sql_ts(finished),
                sql_num(random.randint(100, 5000)),
                sql_num(random.randint(0, 30)),
                sql_str(None if status != "failed" else "Ошибка загрузки части данных"),
                sql_ts(started),
            ],
        )
    w.line()

    w.comment("История job-ов анализа")
    for i in range(1, 6):
        started = rand_dt(30)
        finished = started + timedelta(minutes=random.randint(1, 25))
        status = pick(JOB_STATUSES)
        w.insert(
            "analysis_jobs",
            [
                "id",
                "job_type",
                "status",
                "started_at",
                "finished_at",
                "sellers_processed",
                "recommendations_created",
                "triggers_created",
                "error_message",
                "created_at",
            ],
            [
                sql_num(i),
                sql_str(pick(["daily_analysis", "manual_run", "weekly_scoring"])),
                sql_str(status),
                sql_ts(started),
                sql_ts(finished),
                sql_num(random.randint(100, SELLERS_COUNT)),
                sql_num(random.randint(10, SELLERS_COUNT)),
                sql_num(random.randint(10, SELLERS_COUNT * 2)),
                sql_str(None if status != "failed" else "Ошибка анализа части продавцов"),
                sql_ts(started),
            ],
        )
    w.line()


def write_sellers_and_related(w: SqlWriter, reg: Registry, rec_type_map: dict[str, int]) -> None:
    product_id_seq = 1
    order_id_seq = 1
    activity_id_seq = 1
    category_presence_id_seq = 1
    snapshot_id_seq = 1
    category_metric_id_seq = 1
    region_metric_id_seq = 1
    feature_id_seq = 1
    trigger_id_seq = 1
    recommendation_id_seq = 1
    notification_id_seq = 1
    feedback_id_seq = 1
    outbox_id_seq = 1
    seller_model_id_seq = 1

    template_ids_by_rec_type: dict[int, list[int]] = {}
    tpl_counter = 1
    for _, _, _ in RECOMMENDATION_TYPE_DATA:
        pass
    # templates already generated in order: one rec type x channels
    for rec_type_id in reg.recommendation_types:
        template_ids_by_rec_type[rec_type_id] = [tpl_counter, tpl_counter + 1, tpl_counter + 2]
        tpl_counter += 3

    rule_to_rec_type: dict[int, int] = {}
    for idx, rule_data in enumerate(RULE_DATA, start=1):
        _, _, _, rec_code = rule_data
        rule_to_rec_type[idx] = rec_type_map[rec_code]

    for seller_id in range(1, SELLERS_COUNT + 1):
        reg.sellers.append(seller_id)

        registration_at = rand_dt(365)
        last_login_at = registration_at + timedelta(days=random.randint(0, 180))
        if last_login_at > NOW:
            last_login_at = NOW - timedelta(days=random.randint(0, 5))

        seller_status = pick(SELLER_STATUSES)
        lifecycle_stage = pick(LIFECYCLE_STAGES)
        home_region_id = pick(reg.regions)

        # sellers
        w.insert(
            "sellers",
            [
                "id",
                "external_seller_id",
                "seller_name",
                "seller_type",
                "status",
                "lifecycle_stage",
                "registration_at",
                "last_login_at",
                "home_region_id",
                "created_at",
                "updated_at",
            ],
            [
                sql_num(seller_id),
                sql_str(f"EXT_SELLER_{seller_id:05d}"),
                sql_str(f"Продавец {seller_id}"),
                sql_str(pick(SELLER_TYPES)),
                sql_str(seller_status),
                sql_str(lifecycle_stage),
                sql_ts(registration_at),
                sql_ts(last_login_at),
                sql_num(home_region_id),
                sql_ts(registration_at),
                sql_ts(rand_dt(30)),
            ],
        )

        # seller_models
        model_count = random.randint(SELLER_MODELS_PER_SELLER_MIN, SELLER_MODELS_PER_SELLER_MAX)
        for model_code in random.sample(MODEL_CODES, k=model_count):
            enabled_at = rand_dt(180)
            disabled_at = None
            status = pick(MODEL_STATUSES)
            if status == "disabled":
                disabled_at = enabled_at + timedelta(days=random.randint(1, 60))
            w.insert(
                "seller_models",
                [
                    "id",
                    "seller_id",
                    "model_code",
                    "status",
                    "enabled_at",
                    "disabled_at",
                    "created_at",
                    "updated_at",
                ],
                [
                    sql_num(seller_model_id_seq),
                    sql_num(seller_id),
                    sql_str(model_code),
                    sql_str(status),
                    sql_ts(enabled_at),
                    sql_ts(disabled_at),
                    sql_ts(enabled_at),
                    sql_ts(rand_dt(30)),
                ],
            )
            seller_model_id_seq += 1

        # seller_products
        seller_product_ids: list[int] = []
        product_count = random.randint(PRODUCTS_PER_SELLER_MIN, PRODUCTS_PER_SELLER_MAX)
        used_categories: set[int] = set()
        for _ in range(product_count):
            category_id = pick(reg.categories)
            used_categories.add(category_id)
            status = pick(PRODUCT_STATUSES)
            published_at = rand_dt(180) if status in {"active", "inactive", "archived"} else None
            unpublished_at = rand_dt(60) if status in {"inactive", "archived"} else None
            price = money(200, 12000)
            stock_qty = random.randint(0, 120)

            reg.seller_products.append(product_id_seq)
            seller_product_ids.append(product_id_seq)

            w.insert(
                "seller_products",
                [
                    "id",
                    "seller_id",
                    "external_product_id",
                    "sku",
                    "product_name",
                    "category_id",
                    "status",
                    "price",
                    "stock_qty",
                    "published_at",
                    "unpublished_at",
                    "created_at",
                    "updated_at",
                    "loaded_at",
                ],
                [
                    sql_num(product_id_seq),
                    sql_num(seller_id),
                    sql_str(f"EXT_PRODUCT_{product_id_seq:06d}"),
                    sql_str(f"SKU-{seller_id:05d}-{product_id_seq:06d}"),
                    sql_str(f"Товар {product_id_seq} продавца {seller_id}"),
                    sql_num(category_id),
                    sql_str(status),
                    sql_num(price),
                    sql_num(stock_qty),
                    sql_ts(published_at),
                    sql_ts(unpublished_at),
                    sql_ts(rand_dt(180)),
                    sql_ts(rand_dt(30)),
                    sql_ts(rand_dt(10)),
                ],
            )
            product_id_seq += 1

        # seller_category_presence
        for category_id in used_categories:
            first_seen_at = rand_dt(180)
            last_seen_at = first_seen_at + timedelta(days=random.randint(0, 180))
            if last_seen_at > NOW:
                last_seen_at = NOW
            w.insert(
                "seller_category_presence",
                [
                    "id",
                    "seller_id",
                    "category_id",
                    "first_seen_at",
                    "last_seen_at",
                    "is_active",
                ],
                [
                    sql_num(category_presence_id_seq),
                    sql_num(seller_id),
                    sql_num(category_id),
                    sql_ts(first_seen_at),
                    sql_ts(last_seen_at),
                    sql_bool(chance(0.8)),
                ],
            )
            category_presence_id_seq += 1

        # seller_orders
        order_count = random.randint(ORDERS_PER_SELLER_MIN, ORDERS_PER_SELLER_MAX)
        seller_region_ids_used: set[int] = set()
        for _ in range(order_count):
            product_id = pick(seller_product_ids) if seller_product_ids and chance(0.9) else None
            region_id = pick(reg.regions)
            seller_region_ids_used.add(region_id)

            ordered_at = rand_dt(180)
            order_status = pick(ORDER_STATUSES)
            completed_at = None
            cancelled_at = None

            if order_status == "completed":
                completed_at = ordered_at + timedelta(days=random.randint(1, 12))
            elif order_status == "cancelled":
                cancelled_at = ordered_at + timedelta(days=random.randint(0, 5))

            amount = money(300, 20000)
            margin = round(amount * random.uniform(0.05, 0.35), 2)

            reg.seller_orders.append(order_id_seq)

            w.insert(
                "seller_orders",
                [
                    "id",
                    "seller_id",
                    "external_order_id",
                    "product_id",
                    "region_id",
                    "order_status",
                    "order_amount",
                    "margin_amount",
                    "ordered_at",
                    "completed_at",
                    "cancelled_at",
                    "created_at",
                    "updated_at",
                    "loaded_at",
                ],
                [
                    sql_num(order_id_seq),
                    sql_num(seller_id),
                    sql_str(f"EXT_ORDER_{order_id_seq:07d}"),
                    sql_num(product_id),
                    sql_num(region_id),
                    sql_str(order_status),
                    sql_num(amount),
                    sql_num(margin),
                    sql_ts(ordered_at),
                    sql_ts(completed_at),
                    sql_ts(cancelled_at),
                    sql_ts(ordered_at),
                    sql_ts(rand_dt(30)),
                    sql_ts(rand_dt(10)),
                ],
            )
            order_id_seq += 1

        # seller_activity_log
        activity_count = random.randint(ACTIVITY_PER_SELLER_MIN, ACTIVITY_PER_SELLER_MAX)
        for _ in range(activity_count):
            activity_at = rand_dt(120)
            payload = {
                "screen": pick(["dashboard", "products", "analytics", "orders"]),
                "source": "seller_cabinet",
            }
            w.insert(
                "seller_activity_log",
                ["id", "seller_id", "activity_type", "activity_at", "source_system", "payload_json", "loaded_at"],
                [
                    sql_num(activity_id_seq),
                    sql_num(seller_id),
                    sql_str(pick(ACTIVITY_TYPES)),
                    sql_ts(activity_at),
                    sql_str("seller_cabinet"),
                    sql_json(payload),
                    sql_ts(rand_dt(5)),
                ],
            )
            activity_id_seq += 1

        # seller_metrics_snapshot
        seller_snapshot_ids: list[int] = []
        for day_shift in range(SNAPSHOTS_DAYS):
            snapshot_date = (NOW - timedelta(days=day_shift)).date()

            active_products_count = random.randint(0, max(1, product_count))
            published_products_count = random.randint(active_products_count, product_count) if active_products_count <= product_count else product_count
            products_without_stock_count = random.randint(0, product_count)
            categories_count = len(used_categories)
            active_categories_count = random.randint(0, categories_count) if categories_count > 0 else 0
            regions_count = len(seller_region_ids_used) if seller_region_ids_used else random.randint(1, 3)
            orders_7d = random.randint(0, 15)
            orders_30d = random.randint(orders_7d, 30)
            revenue_7d = round(orders_7d * money(200, 2000), 2)
            revenue_30d = round(orders_30d * money(200, 2000), 2)
            margin_30d = round(revenue_30d * random.uniform(0.05, 0.3), 2)
            last_login_days = random.randint(0, 30)
            no_sales_days = random.randint(0, 45)
            current_primary_model_code = pick(MODEL_CODES)

            reg.seller_snapshots.append(snapshot_id_seq)
            seller_snapshot_ids.append(snapshot_id_seq)

            w.insert(
                "seller_metrics_snapshot",
                [
                    "id",
                    "seller_id",
                    "snapshot_date",
                    "active_products_count",
                    "published_products_count",
                    "products_without_stock_count",
                    "categories_count",
                    "active_categories_count",
                    "regions_count",
                    "orders_7d",
                    "orders_30d",
                    "revenue_7d",
                    "revenue_30d",
                    "margin_30d",
                    "last_login_days",
                    "no_sales_days",
                    "current_primary_model_code",
                    "created_at",
                ],
                [
                    sql_num(snapshot_id_seq),
                    sql_num(seller_id),
                    sql_date(snapshot_date),
                    sql_num(active_products_count),
                    sql_num(published_products_count),
                    sql_num(products_without_stock_count),
                    sql_num(categories_count),
                    sql_num(active_categories_count),
                    sql_num(regions_count),
                    sql_num(orders_7d),
                    sql_num(orders_30d),
                    sql_num(revenue_7d),
                    sql_num(revenue_30d),
                    sql_num(margin_30d),
                    sql_num(last_login_days),
                    sql_num(no_sales_days),
                    sql_str(current_primary_model_code),
                    sql_ts(rand_dt(5)),
                ],
            )
            snapshot_id_seq += 1

        # seller_category_metrics
        metric_categories = random.sample(list(used_categories), k=min(len(used_categories), random.randint(CATEGORY_METRICS_PER_SELLER_MIN, max(CATEGORY_METRICS_PER_SELLER_MIN, min(len(used_categories), CATEGORY_METRICS_PER_SELLER_MAX))))) if used_categories else []
        for category_id in metric_categories:
            metric_date = rand_date(30)
            orders_count = random.randint(0, 20)
            revenue = round(orders_count * money(100, 2000), 2)
            margin = round(revenue * random.uniform(0.05, 0.3), 2)
            last_sale_at = rand_dt(30) if orders_count > 0 else None
            w.insert(
                "seller_category_metrics",
                [
                    "id",
                    "seller_id",
                    "category_id",
                    "metric_date",
                    "products_count",
                    "orders_count",
                    "revenue",
                    "margin",
                    "last_sale_at",
                    "created_at",
                ],
                [
                    sql_num(category_metric_id_seq),
                    sql_num(seller_id),
                    sql_num(category_id),
                    sql_date(metric_date),
                    sql_num(random.randint(1, 10)),
                    sql_num(orders_count),
                    sql_num(revenue),
                    sql_num(margin),
                    sql_ts(last_sale_at),
                    sql_ts(rand_dt(5)),
                ],
            )
            category_metric_id_seq += 1

        # seller_region_metrics
        region_candidates = list(seller_region_ids_used) if seller_region_ids_used else random.sample(reg.regions, k=min(3, len(reg.regions)))
        k = min(len(region_candidates), random.randint(REGION_METRICS_PER_SELLER_MIN, max(REGION_METRICS_PER_SELLER_MIN, min(len(region_candidates), REGION_METRICS_PER_SELLER_MAX))))
        for region_id in random.sample(region_candidates, k=k):
            metric_date = rand_date(30)
            orders_count = random.randint(0, 20)
            revenue = round(orders_count * money(100, 2500), 2)
            margin = round(revenue * random.uniform(0.05, 0.3), 2)
            avg_delivery_days = round(random.uniform(1.0, 12.0), 2)
            w.insert(
                "seller_region_metrics",
                [
                    "id",
                    "seller_id",
                    "region_id",
                    "metric_date",
                    "orders_count",
                    "revenue",
                    "margin",
                    "avg_delivery_days",
                    "created_at",
                ],
                [
                    sql_num(region_metric_id_seq),
                    sql_num(seller_id),
                    sql_num(region_id),
                    sql_date(metric_date),
                    sql_num(orders_count),
                    sql_num(revenue),
                    sql_num(margin),
                    sql_num(avg_delivery_days),
                    sql_ts(rand_dt(5)),
                ],
            )
            region_metric_id_seq += 1

        # seller_features
        for idx in range(FEATURES_PER_SELLER):
            feature_date = rand_date(7)
            features_json = {
                "active_products_ratio": round(random.uniform(0.0, 1.0), 4),
                "orders_trend_30d": round(random.uniform(-1.0, 1.0), 4),
                "login_frequency_score": round(random.uniform(0.0, 1.0), 4),
                "seller_cluster": pick(["new", "stable", "growing", "at_risk"]),
            }
            w.insert(
                "seller_features",
                ["id", "seller_id", "feature_date", "feature_version", "features_json", "created_at"],
                [
                    sql_num(feature_id_seq),
                    sql_num(seller_id),
                    sql_date(feature_date),
                    sql_str(f"v1.{idx}"),
                    sql_json(features_json),
                    sql_ts(rand_dt(5)),
                ],
            )
            feature_id_seq += 1

        # triggers + recommendations + notifications + feedback + outbox
        recommendations_count = random.randint(RECOMMENDATIONS_PER_SELLER_MIN, RECOMMENDATIONS_PER_SELLER_MAX)
        used_rule_periods: set[tuple[int, str]] = set()

        for _ in range(recommendations_count):
            rule_id = pick(reg.rules)
            period_key = f"{NOW.year}-W{random.randint(1, 52):02d}"
            while (rule_id, period_key) in used_rule_periods:
                period_key = f"{NOW.year}-W{random.randint(1, 52):02d}"
            used_rule_periods.add((rule_id, period_key))

            trigger_code = pick(TRIGGER_CODES)
            snapshot_id = pick(seller_snapshot_ids) if seller_snapshot_ids else None
            trigger_status = pick(["detected", "converted_to_recommendation"])

            reg.triggers.append(trigger_id_seq)

            trigger_payload = {
                "detected_reason": trigger_code,
                "source": "analysis_service",
                "score_hint": round(random.uniform(0.1, 0.99), 4),
            }

            trigger_time = rand_dt(20)
            w.insert(
                "seller_trigger_log",
                [
                    "id",
                    "seller_id",
                    "rule_id",
                    "trigger_code",
                    "triggered_at",
                    "period_key",
                    "snapshot_id",
                    "payload_json",
                    "status",
                    "created_at",
                ],
                [
                    sql_num(trigger_id_seq),
                    sql_num(seller_id),
                    sql_num(rule_id),
                    sql_str(trigger_code),
                    sql_ts(trigger_time),
                    sql_str(period_key),
                    sql_num(snapshot_id),
                    sql_json(trigger_payload),
                    sql_str(trigger_status),
                    sql_ts(trigger_time),
                ],
            )

            rec_type_id = rule_to_rec_type[rule_id]
            template_id = pick(template_ids_by_rec_type[rec_type_id])
            recommendation_status = pick(RECOMMENDATION_STATUSES)
            rec_created_at = trigger_time + timedelta(minutes=random.randint(1, 120))
            score = round(random.uniform(0.1, 0.99), 4)

            reg.recommendations.append(recommendation_id_seq)

            w.insert(
                "recommendations",
                [
                    "id",
                    "seller_id",
                    "trigger_id",
                    "recommendation_type_id",
                    "template_id",
                    "title",
                    "description",
                    "reason_text",
                    "priority",
                    "score",
                    "status",
                    "expires_at",
                    "created_at",
                    "updated_at",
                ],
                [
                    sql_num(recommendation_id_seq),
                    sql_num(seller_id),
                    sql_num(trigger_id_seq),
                    sql_num(rec_type_id),
                    sql_num(template_id),
                    sql_str(f"Рекомендация {recommendation_id_seq}"),
                    sql_str(f"Описание рекомендации {recommendation_id_seq} для продавца {seller_id}"),
                    sql_str(f"Сработало правило {rule_id} и триггер {trigger_code}"),
                    sql_num(random.randint(50, 300)),
                    sql_num(score),
                    sql_str(recommendation_status),
                    sql_ts(rec_created_at + timedelta(days=random.randint(7, 30))),
                    sql_ts(rec_created_at),
                    sql_ts(rand_dt(5)),
                ],
            )

            # notifications
            notif_count = random.randint(1, MAX_NOTIFICATIONS_PER_RECOMMENDATION)
            for _n in range(notif_count):
                notif_status = pick(NOTIFICATION_STATUSES)
                notif_created = rec_created_at + timedelta(minutes=random.randint(1, 30))
                sent_at = notif_created if notif_status in {"sent", "delivered", "opened", "clicked"} else None
                delivered_at = sent_at + timedelta(minutes=random.randint(1, 60)) if notif_status in {"delivered", "opened", "clicked"} and sent_at else None
                opened_at = delivered_at + timedelta(minutes=random.randint(1, 120)) if notif_status in {"opened", "clicked"} and delivered_at else None
                clicked_at = opened_at + timedelta(minutes=random.randint(1, 60)) if notif_status == "clicked" and opened_at else None

                payload = {
                    "seller_id": seller_id,
                    "recommendation_id": recommendation_id_seq,
                    "channel": pick(CHANNEL_CODES),
                }

                reg.notifications.append(notification_id_seq)

                w.insert(
                    "seller_notification_log",
                    [
                        "id",
                        "seller_id",
                        "recommendation_id",
                        "channel_code",
                        "delivery_system_id",
                        "status",
                        "payload_json",
                        "sent_at",
                        "delivered_at",
                        "opened_at",
                        "clicked_at",
                        "error_message",
                        "created_at",
                        "updated_at",
                    ],
                    [
                        sql_num(notification_id_seq),
                        sql_num(seller_id),
                        sql_num(recommendation_id_seq),
                        sql_str(payload["channel"]),
                        sql_str(f"DELIVERY_{notification_id_seq:07d}" if notif_status != "created" else None),
                        sql_str(notif_status),
                        sql_json(payload),
                        sql_ts(sent_at),
                        sql_ts(delivered_at),
                        sql_ts(opened_at),
                        sql_ts(clicked_at),
                        sql_str("Ошибка доставки" if notif_status == "failed" else None),
                        sql_ts(notif_created),
                        sql_ts(rand_dt(3)),
                    ],
                )

                # outbox
                outbox_status = pick(OUTBOX_STATUSES)
                attempt_count = random.randint(0, MAX_OUTBOX_RETRIES)
                next_retry_at = rand_dt(3) if outbox_status != "sent" and chance(0.5) else None

                w.insert(
                    "outbound_notification_queue",
                    [
                        "id",
                        "recommendation_id",
                        "notification_id",
                        "status",
                        "attempt_count",
                        "next_retry_at",
                        "last_error",
                        "created_at",
                        "updated_at",
                    ],
                    [
                        sql_num(outbox_id_seq),
                        sql_num(recommendation_id_seq),
                        sql_num(notification_id_seq),
                        sql_str(outbox_status),
                        sql_num(attempt_count),
                        sql_ts(next_retry_at),
                        sql_str("Повторная ошибка отправки" if outbox_status == "failed" else None),
                        sql_ts(notif_created),
                        sql_ts(rand_dt(2)),
                    ],
                )
                outbox_id_seq += 1
                notification_id_seq += 1

            # feedback
            feedback_count = random.randint(0, MAX_FEEDBACK_PER_RECOMMENDATION)
            for _f in range(feedback_count):
                feedback_type = pick(FEEDBACK_TYPES)
                feedback_at = rec_created_at + timedelta(minutes=random.randint(5, 1440))
                payload = {"source": "seller_ui", "action": feedback_type}
                w.insert(
                    "recommendation_feedback",
                    [
                        "id",
                        "seller_id",
                        "recommendation_id",
                        "feedback_type",
                        "feedback_at",
                        "payload_json",
                        "created_at",
                    ],
                    [
                        sql_num(feedback_id_seq),
                        sql_num(seller_id),
                        sql_num(recommendation_id_seq),
                        sql_str(feedback_type),
                        sql_ts(feedback_at),
                        sql_json(payload),
                        sql_ts(feedback_at),
                    ],
                )
                feedback_id_seq += 1

            trigger_id_seq += 1
            recommendation_id_seq += 1

    w.line()


def write_footer(w: SqlWriter) -> None:
    w.line("commit;")
    w.line()


# =========================================================
# MAIN
# =========================================================

def main() -> None:
    w = SqlWriter()
    reg = Registry()

    write_header(w)
    write_cleanup(w)
    write_regions(w, reg)
    write_categories(w, reg)
    rec_type_map = write_recommendation_types(w, reg)
    write_rules(w, reg, rec_type_map)
    write_templates(w, reg, rec_type_map)
    write_jobs(w)
    write_sellers_and_related(w, reg, rec_type_map)
    write_footer(w)

    w.save(OUTPUT_FILE)

    print("Готово.")
    print(f"SQL-файл создан: {OUTPUT_FILE}")
    print(f"Продавцов: {SELLERS_COUNT}")
    print(f"Регионов: {REGIONS_COUNT}")
    print(f"Категорий: {CATEGORIES_COUNT}")
    print("Объем связанных таблиц рассчитан автоматически.")


if __name__ == "__main__":
    main()