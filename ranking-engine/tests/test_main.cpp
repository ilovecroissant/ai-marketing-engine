#include "cosine.hpp"
#include "ranking.hpp"
#include "tfidf.hpp"
#include <algorithm>
#include <cassert>
#include <cmath>
#include <iostream>
#include <string>

// ─── helpers ────────────────────────────────────────────────────────────────

static int g_passed = 0;
static int g_failed = 0;

#define ASSERT_TRUE(cond)                                                    \
    do {                                                                     \
        if (!(cond)) {                                                       \
            std::cerr << "  FAIL " << __FILE__ << ":" << __LINE__           \
                      << "  " #cond "\n";                                    \
            ++g_failed; return;                                              \
        }                                                                    \
    } while (0)

#define ASSERT_NEAR(a, b, tol)                                               \
    do {                                                                     \
        if (std::abs((a) - (b)) > (tol)) {                                  \
            std::cerr << "  FAIL " << __FILE__ << ":" << __LINE__           \
                      << "  expected ~" << (b) << " got " << (a) << "\n";   \
            ++g_failed; return;                                              \
        }                                                                    \
    } while (0)

#define TEST(name)                                                           \
    static void name();                                                      \
    struct _reg_##name { _reg_##name() {                                     \
        std::cout << "  RUN  " #name "\n";                                   \
        int before = g_failed;                                               \
        name();                                                               \
        if (g_failed == before) { ++g_passed; std::cout << "  PASS\n"; }    \
        else                    { std::cout << "  FAIL\n"; }                 \
    }} _inst_##name;                                                         \
    static void name()

// ─── tokenise ───────────────────────────────────────────────────────────────

TEST(tokenise_strips_punctuation_and_stop_words) {
    auto tokens = ranking::tokenise("Hello, World! This is a test.");
    bool hasHello = std::find(tokens.begin(), tokens.end(), "hello") != tokens.end();
    bool hasWorld = std::find(tokens.begin(), tokens.end(), "world") != tokens.end();
    bool hasTest  = std::find(tokens.begin(), tokens.end(), "test")  != tokens.end();
    // "this" "is" "a" should be removed as stop words
    bool hasThis  = std::find(tokens.begin(), tokens.end(), "this")  != tokens.end();
    ASSERT_TRUE(hasHello && hasWorld && hasTest);
    ASSERT_TRUE(!hasThis);
}

// ─── TF ─────────────────────────────────────────────────────────────────────

TEST(tf_counts_are_proportional) {
    // "apple" 3×, "banana" 2× → raw tokens = 5
    auto tf = ranking::computeTF("apple apple apple banana banana");
    ASSERT_NEAR(tf.at("apple"),  0.6, 0.01);
    ASSERT_NEAR(tf.at("banana"), 0.4, 0.01);
}

TEST(tf_empty_text_returns_empty) {
    auto tf = ranking::computeTF("");
    ASSERT_TRUE(tf.empty());
}

// ─── IDF ────────────────────────────────────────────────────────────────────

TEST(idf_rare_terms_score_higher) {
    std::vector<std::string> corpus = {
        "apple banana cherry",
        "apple date elderberry",
        "fig grape honeydew"
    };
    auto idf = ranking::computeIDF(corpus);
    // "apple" in 2 docs, "fig" in 1 doc → fig should have higher IDF
    ASSERT_TRUE(idf.count("apple") && idf.count("fig"));
    ASSERT_TRUE(idf.at("fig") > idf.at("apple"));
}

TEST(idf_empty_corpus_returns_empty) {
    ASSERT_TRUE(ranking::computeIDF({}).empty());
}

// ─── TF-IDF ─────────────────────────────────────────────────────────────────

TEST(tfidf_term_order_reflects_uniqueness) {
    std::vector<std::string> corpus = {
        "machine learning algorithms models training",
        "machine learning neural networks deep",
        "cooking recipes ingredients kitchen food"
    };
    auto idf   = ranking::computeIDF(corpus);
    auto tfidf = ranking::computeTFIDF("machine learning cooking", idf);

    // "cooking" appears in only 1 doc so its tfidf should be higher than
    // "machine" which appears in 2 docs
    ASSERT_TRUE(tfidf.count("cooking") && tfidf.count("machine"));
    ASSERT_TRUE(tfidf.at("cooking") > tfidf.at("machine"));
}

// ─── top keywords ────────────────────────────────────────────────────────────

TEST(top_keywords_are_sorted_descending) {
    auto tfidf = ranking::computeTFIDFSingle(
        "marketing content strategy digital marketing content creation marketing");
    auto kws = ranking::topKeywords(tfidf, 5);

    ASSERT_TRUE(!kws.empty());
    ASSERT_TRUE(kws.size() <= 5);
    for (size_t i = 1; i < kws.size(); ++i)
        ASSERT_TRUE(kws[i - 1].second >= kws[i].second);
    // "marketing" appears 3× so should be rank 1
    ASSERT_TRUE(kws[0].first == "marketing");
}

TEST(top_keywords_n_larger_than_vocab_clamps) {
    auto tfidf = ranking::computeTFIDFSingle("one two three");
    auto kws   = ranking::topKeywords(tfidf, 100);
    ASSERT_TRUE(kws.size() <= 3); // vocab has at most 3 terms
}

// ─── cosine similarity ───────────────────────────────────────────────────────

TEST(cosine_identical_texts_score_one) {
    std::string t = "machine learning artificial intelligence algorithms";
    ASSERT_NEAR(ranking::textSimilarity(t, t), 1.0, 0.01);
}

TEST(cosine_disjoint_texts_score_zero) {
    double sim = ranking::textSimilarity(
        "apple orange banana fruit salad",
        "quantum physics electrons protons neutrons");
    ASSERT_NEAR(sim, 0.0, 0.05);
}

TEST(cosine_similar_texts_score_above_half) {
    double sim = ranking::textSimilarity(
        "digital marketing content strategy creation",
        "content creation digital marketing strategy");
    ASSERT_TRUE(sim > 0.8);
}

TEST(cosine_empty_input_returns_zero) {
    ASSERT_NEAR(ranking::textSimilarity("", "some text"), 0.0, 0.01);
    ASSERT_NEAR(ranking::textSimilarity("", ""),           0.0, 0.01);
}

// ─── ranking ────────────────────────────────────────────────────────────────

TEST(scores_are_in_unit_interval) {
    std::string content =
        "Digital marketing strategies help businesses grow online. "
        "Content marketing involves creating valuable content for your audience. "
        "SEO optimization improves search engine rankings. "
        "Social media marketing drives engagement and brand awareness. "
        "Email marketing campaigns reach targeted customers effectively. "
        "Paid advertising through PPC marketing delivers measurable results.";

    auto s = ranking::scoreContent(content, {"marketing", "digital", "content", "seo"});
    ASSERT_TRUE(s.seo         >= 0.0 && s.seo         <= 1.0);
    ASSERT_TRUE(s.engagement  >= 0.0 && s.engagement  <= 1.0);
    ASSERT_TRUE(s.readability >= 0.0 && s.readability <= 1.0);
    ASSERT_TRUE(s.total       >= 0.0 && s.total       <= 1.0);
}

TEST(total_equals_weighted_sum) {
    auto s = ranking::scoreContent(
        "marketing digital content seo strategy growth analytics platform",
        {"marketing", "content"});
    double expected = 0.4 * s.seo + 0.3 * s.engagement + 0.3 * s.readability;
    ASSERT_NEAR(s.total, expected, 1e-9);
}

TEST(empty_text_does_not_crash) {
    auto s = ranking::scoreContent("", {"marketing"});
    ASSERT_TRUE(s.total >= 0.0 && s.total <= 1.0);
}

TEST(no_keywords_returns_neutral_seo) {
    auto s = ranking::scoreContent("some content about things", {});
    // SEO defaults to 0.5 when no keywords given
    ASSERT_NEAR(s.seo, 0.5, 0.01);
}

// ─── entry point ─────────────────────────────────────────────────────────────

int main() {
    std::cout << "\n=== Ranking Engine Unit Tests ===\n\n";
    // All TEST blocks self-register and run at static-init time.
    // We just report the final tally here.
    if (g_failed == 0) {
        std::cout << "\nAll " << g_passed << " tests passed.\n";
        return 0;
    }
    std::cout << "\n" << g_failed << " test(s) FAILED, "
              << g_passed << " passed.\n";
    return 1;
}
