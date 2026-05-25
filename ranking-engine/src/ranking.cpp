#include "ranking.hpp"
#include "tfidf.hpp"
#include <algorithm>
#include <cctype>
#include <cmath>

namespace ranking {

// Maps a raw value to [0,1] using a logistic curve centred at `midpoint`.
static double logistic(double value, double midpoint, double scale) {
    return 1.0 / (1.0 + std::exp(-(value - midpoint) / scale));
}

// SEO score: how well the content covers the target keywords.
// Considers both keyword presence (coverage) and density (avoids stuffing).
static double computeSEO(const std::string& text,
                          const std::vector<std::string>& keywords) {
    if (keywords.empty()) return 0.5;

    auto tokens = tokenise(text);
    if (tokens.empty()) return 0.0;

    std::unordered_map<std::string, int> counts;
    for (const auto& t : tokens) counts[t]++;

    int hits = 0;
    double totalDensity = 0.0;
    double total = static_cast<double>(tokens.size());

    for (const auto& kw : keywords) {
        std::string lkw;
        for (unsigned char c : kw) lkw += static_cast<char>(std::tolower(c));
        auto it = counts.find(lkw);
        if (it != counts.end()) {
            hits++;
            totalDensity += it->second / total;
        }
    }

    double coverage    = static_cast<double>(hits) / keywords.size();
    // Ideal density ≈ 2–4 % per keyword; penalise over-stuffing
    double avgDensity  = hits > 0 ? totalDensity / hits : 0.0;
    double densityScore = std::max(0.0, 1.0 - std::abs(avgDensity - 0.03) / 0.03);

    return 0.6 * coverage + 0.4 * densityScore;
}

// Engagement score: content length and sentence-length variety.
static double computeEngagement(const std::string& text) {
    auto tokens = tokenise(text);
    int words = static_cast<int>(tokens.size());

    // Target: 300–1200 words; logistic centred at 600
    double lengthScore = logistic(static_cast<double>(words), 600.0, 150.0);

    int sentences = 0;
    for (unsigned char c : text)
        if (c == '.' || c == '!' || c == '?') sentences++;
    if (sentences == 0) sentences = 1;

    // Ideal: ~15–20 words per sentence
    double avgSentLen = static_cast<double>(words) / sentences;
    double sentScore  = logistic(avgSentLen, 17.0, 4.0);

    return 0.6 * lengthScore + 0.4 * sentScore;
}

// Readability score: average word length and syllable count as proxies for
// reading complexity (lower complexity = higher readability score).
static double computeReadability(const std::string& text) {
    auto tokens = tokenise(text);
    if (tokens.empty()) return 0.5;

    double totalChars = 0.0;
    int vowelGroups   = 0;
    bool inVowel      = false;
    const std::string vowels = "aeiou";

    for (const auto& t : tokens) {
        totalChars += t.size();
        inVowel = false;
        for (char c : t) {
            bool isVowel = vowels.find(c) != std::string::npos;
            if (isVowel && !inVowel) { vowelGroups++; inVowel = true; }
            else if (!isVowel)         { inVowel = false; }
        }
    }

    double n = static_cast<double>(tokens.size());
    double avgWordLen  = totalChars / n;
    double avgSyllable = static_cast<double>(vowelGroups) / n;

    // Ideal word length ~5 chars, syllable count ~1.5–2.0 per word
    double wordScore     = 1.0 - std::min(1.0, std::abs(avgWordLen - 5.0) / 5.0);
    double syllableScore = 1.0 - std::min(1.0, std::abs(avgSyllable - 1.75) / 1.75);

    return 0.5 * wordScore + 0.5 * syllableScore;
}

ContentScore scoreContent(const std::string& text,
                           const std::vector<std::string>& keywords) {
    ContentScore s;
    s.seo         = computeSEO(text, keywords);
    s.engagement  = computeEngagement(text);
    s.readability = computeReadability(text);
    s.total       = 0.4 * s.seo + 0.3 * s.engagement + 0.3 * s.readability;
    return s;
}

} // namespace ranking
