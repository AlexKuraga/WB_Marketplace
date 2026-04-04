workspace "Seller Recommendation Platform" "C2 container diagram with external data source and loader service" {

    model {
        seller = person "Продавец" "Получает рекомендации и уведомления."
        operator = person "Бизнес-оператор" "Управляет правилами и контролирует работу сервиса."

        externalDataSystem = softwareSystem "Внешняя система данных" "Источник данных о продавцах, товарах, продажах, активности и других бизнес-показателях."
        notificationDelivery = softwareSystem "Сервис доставки уведомлений" "Определяет канал, место и момент показа уведомления пользователю."

        sellerRecommendationPlatform = softwareSystem "Seller Recommendation Platform" "Платформа анализа данных продавцов, генерации триггеров и подготовки уведомлений." {
            adminApi = container "Admin API" "API для управления правилами, ручного запуска пересчета и просмотра результатов." "Java/Kotlin + Spring Boot"
            recommendationApi = container "Recommendation API" "Отдает рекомендации в UI и принимает feedback по уведомлениям." "Java/Kotlin + Spring Boot"
            dataLoaderScheduler = container "Data Loader Scheduler" "Запускает периодическую загрузку данных из внешней системы." "Java/Kotlin + Quartz/Cron"
            dataLoaderService = container "Data Loader Service" "Читает данные из внешней системы, валидирует, преобразует и сохраняет в БД." "Java/Kotlin"
            analysisScheduler = container "Analysis Scheduler" "Запускает периодический анализ данных по расписанию." "Java/Kotlin + Quartz/Cron"
            analysisService = container "Analysis Service" "Читает данные продавцов из БД, рассчитывает признаки и формирует кандидаты на триггеры." "Java/Kotlin"
            ruleEngine = container "Rule Engine" "Проверяет бизнес-правила, дедупликацию и cooldown, принимает решение о создании уведомления." "Java/Kotlin"
            notificationHandoff = container "Notification Handoff Adapter" "Передает готовые уведомления в сервис доставки и обновляет статусы интеграции." "Java/Kotlin"
            postgres = container "Operational Database" "Единая БД: sellers, seller_data, seller_metrics, recommendation_rules, seller_trigger_log, seller_notification_log, feedback." "PostgreSQL"
        }

        operator -> adminApi "Управляет правилами и запускает операции" "HTTPS"
        seller -> recommendationApi "Получает рекомендации / отправляет feedback" "HTTPS"

        dataLoaderScheduler -> dataLoaderService "Запускает периодическую загрузку" "Internal"
        dataLoaderService -> externalDataSystem "Запрашивает данные" "HTTPS/REST"
        dataLoaderService -> postgres "Сохраняет загруженные данные" "JDBC"

        analysisScheduler -> analysisService "Запускает периодический анализ" "Internal"
        analysisService -> postgres "Читает данные продавцов и метрики" "JDBC"
        analysisService -> ruleEngine "Передает рассчитанные признаки и кандидаты триггеров" "Internal"

        ruleEngine -> postgres "Читает правила, проверяет trigger log / notification log, пишет новые trigger и notification records" "JDBC"
        ruleEngine -> notificationHandoff "Передает уведомления в статусе READY_TO_SEND" "Internal"

        notificationHandoff -> notificationDelivery "Создает задачу на доставку уведомления" "HTTPS/JSON"
        notificationHandoff -> postgres "Обновляет статусы отправки" "JDBC"

        recommendationApi -> postgres "Читает рекомендации и пишет feedback пользователя" "JDBC"
        adminApi -> postgres "Управляет правилами и конфигурацией" "JDBC"
    }

    views {
        container sellerRecommendationPlatform "C2" {
            include seller
            include operator
            include externalDataSystem
            include notificationDelivery
            include adminApi
            include recommendationApi
            include dataLoaderScheduler
            include dataLoaderService
            include analysisScheduler
            include analysisService
            include ruleEngine
            include notificationHandoff
            include postgres
            autolayout lr
        }

        styles {
            element "Person" {
                shape person
                background #08427b
                color #ffffff
            }

            element "Software System" {
                background #1168bd
                color #ffffff
            }

            element "Container" {
                background #438dd5
                color #ffffff
            }
        }
    }
}