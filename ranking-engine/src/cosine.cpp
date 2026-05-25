#include "cosine.hpp"
#include "tfidf.hpp"
#include <cmath>

namespace ranking {

double cosineSimilarity(
    const std::unordered_map<std::string, double>& a,
    const std::unordered_map<std::string, double>& b) {
    if (a.empty() || b.empty()) return 0.0;

    double dot = 0.0, normA = 0.0, normB = 0.0;

    for (const auto& [term, val] : a) {
        normA += val * val;
        auto it = b.find(term);
        if (it != b.end()) dot += val * it->second;
    }
    for (const auto& [term, val] : b)
        normB += val * val;

    if (normA == 0.0 || normB == 0.0) return 0.0;
    return dot / (std::sqrt(normA) * std::sqrt(normB));
}

double textSimilarity(const std::string& a, const std::string& b) {
    return cosineSimilarity(computeTFIDFSingle(a), computeTFIDFSingle(b));
}

} // namespace ranking
