workspace "Recommendation System" "Dynamic diagrams for API endpoints" {

    model {

        seller = person "Seller" "Marketplace seller"
        admin = person "Admin" "System administrator"

        delivery_service = softwareSystem "Delivery Service" "External notification delivery service"

        recommendationSystem = softwareSystem "Recommendation System" {

            recommendation_api = container "Recommendation API" "Handles seller recommendations and feedback" "Spring Boot"
            admin_api = container "Admin API" "Administrative operations and rule management" "Spring Boot"
            analysis_service = container "Analysis Service" "Analyzes seller data and prepares trigger candidates" "Python"
            rule_engine = container "Rule Engine" "Evaluates business rules, deduplication and cooldown" "Java"
            notification_adapter = container "Notification Adapter" "Transfers prepared notifications to downstream delivery service" "Java"
            operational_db = container "Operational DB" "Stores sellers, rules, triggers, recommendations, notifications and feedback" "PostgreSQL"
        }

        seller -> recommendation_api "Uses"
        admin -> admin_api "Manages"

        recommendation_api -> operational_db "Reads/Writes"
        admin_api -> operational_db "Reads/Writes"

        analysis_service -> operational_db "Reads seller data and metrics"
        analysis_service -> rule_engine "Sends trigger candidates"
        admin_api -> analysis_service "Triggers analysis"

        rule_engine -> operational_db "Creates triggers and recommendations"
        rule_engine -> notification_adapter "Transfers ready notifications"

        notification_adapter -> delivery_service "Sends notification payload"
        notification_adapter -> operational_db "Updates notification delivery status"
    }

    views {

        systemContext recommendationSystem {
            include *
            autolayout lr
        }

        container recommendationSystem {
            include *
            autolayout lr
        }

        dynamic recommendationSystem "GetSellerRecommendations" {

            seller -> recommendation_api "GET /api/v1/sellers/{sellerId}/recommendations"

            recommendation_api -> operational_db "Load active recommendations for seller"

            operational_db -> recommendation_api "Return recommendations"

            recommendation_api -> seller "Response with recommendations"
        }

        dynamic recommendationSystem "ViewRecommendation" {

            seller -> recommendation_api "POST /api/v1/recommendations/{id}/view"

            recommendation_api -> operational_db "Update recommendation/view status"

            recommendation_api -> operational_db "Save feedback type=view"

            recommendation_api -> seller "200 OK"
        }

        dynamic recommendationSystem "AcceptRecommendation" {

            seller -> recommendation_api "POST /api/v1/recommendations/{id}/accept"

            recommendation_api -> operational_db "Update recommendation status=accepted"

            recommendation_api -> operational_db "Save feedback type=accept"

            recommendation_api -> seller "200 OK"
        }

        dynamic recommendationSystem "RejectRecommendation" {

            seller -> recommendation_api "POST /api/v1/recommendations/{id}/reject"

            recommendation_api -> operational_db "Update recommendation status=rejected"

            recommendation_api -> operational_db "Save feedback type=reject"

            recommendation_api -> seller "200 OK"
        }

        dynamic recommendationSystem "GetRules" {

            admin -> admin_api "GET /api/v1/admin/rules"

            admin_api -> operational_db "Load recommendation rules"

            operational_db -> admin_api "Rules"

            admin_api -> admin "Response with rules"
        }

        dynamic recommendationSystem "CreateRule" {

            admin -> admin_api "POST /api/v1/admin/rules"

            admin_api -> operational_db "Insert new rule into recommendation_rules"

            operational_db -> admin_api "OK"

            admin_api -> admin "Rule created"
        }

        dynamic recommendationSystem "RunAnalysis" {

            admin -> admin_api "POST /api/v1/admin/run-analysis"

            admin_api -> analysis_service "Trigger manual analysis"

            analysis_service -> operational_db "Load sellers, products, orders, activity and metrics"

            analysis_service -> rule_engine "Send trigger candidates"

            rule_engine -> operational_db "Create seller_trigger_log records"

            rule_engine -> operational_db "Create recommendations"

            rule_engine -> operational_db "Create seller_notification_log records"

            rule_engine -> notification_adapter "Transfer notifications ready to send"

            notification_adapter -> delivery_service "Send prepared notification payload"

            notification_adapter -> operational_db "Update notification sending status"
        }

        theme default
    }
}