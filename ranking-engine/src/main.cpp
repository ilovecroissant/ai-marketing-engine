#include <httplib.h>
#include <nlohmann/json.hpp>
#include "cosine.hpp"
#include "ranking.hpp"
#include "tfidf.hpp"
#include <iostream>

using json = nlohmann::json;

static void setJSON(httplib::Response& res, const json& body, int status = 200) {
    res.status = status;
    res.set_content(body.dump(), "application/json");
}

int main() {
    httplib::Server svr;

    // GET /health
    svr.Get("/health", [](const httplib::Request&, httplib::Response& res) {
        setJSON(res, {{"status", "ok"}});
    });

    // POST /keywords
    // Request:  {"text": "...", "n": 10}
    // Response: {"keywords": [{"term": "marketing", "score": 0.42}, ...]}
    svr.Post("/keywords", [](const httplib::Request& req, httplib::Response& res) {
        try {
            auto body = json::parse(req.body);
            std::string text = body.value("text", "");
            int n            = body.value("n", 10);

            auto tfidf = ranking::computeTFIDFSingle(text);
            auto kws   = ranking::topKeywords(tfidf, n);

            json keywords = json::array();
            for (const auto& [term, score] : kws)
                keywords.push_back({{"term", term}, {"score", score}});

            setJSON(res, {{"keywords", keywords}});
        } catch (const std::exception& e) {
            setJSON(res, {{"error", e.what()}}, 400);
        }
    });

    // POST /rank
    // Request:  {"text": "...", "keywords": ["seo", "content"]}
    // Response: {"seo": 0.7, "engagement": 0.6, "readability": 0.8, "total": 0.69}
    svr.Post("/rank", [](const httplib::Request& req, httplib::Response& res) {
        try {
            auto body = json::parse(req.body);
            std::string text = body.value("text", "");
            std::vector<std::string> keywords;
            if (body.contains("keywords"))
                keywords = body["keywords"].get<std::vector<std::string>>();

            auto s = ranking::scoreContent(text, keywords);
            setJSON(res, {
                {"seo",         s.seo},
                {"engagement",  s.engagement},
                {"readability", s.readability},
                {"total",       s.total}
            });
        } catch (const std::exception& e) {
            setJSON(res, {{"error", e.what()}}, 400);
        }
    });

    // POST /similarity
    // Request:  {"a": "...", "b": "..."}
    // Response: {"similarity": 0.91, "is_duplicate": true}
    svr.Post("/similarity", [](const httplib::Request& req, httplib::Response& res) {
        try {
            auto body = json::parse(req.body);
            std::string a = body.value("a", "");
            std::string b = body.value("b", "");

            double sim = ranking::textSimilarity(a, b);
            setJSON(res, {
                {"similarity",   sim},
                {"is_duplicate", sim > 0.85}
            });
        } catch (const std::exception& e) {
            setJSON(res, {{"error", e.what()}}, 400);
        }
    });

    constexpr int PORT = 8082;
    std::cout << "Ranking engine listening on :" << PORT << "\n";
    svr.listen("0.0.0.0", PORT);
}
