#pragma once
#include <string>
#include <unordered_map>

namespace ranking {

// Cosine similarity between two TF-IDF vectors. Returns a value in [0, 1].
double cosineSimilarity(
    const std::unordered_map<std::string, double>& a,
    const std::unordered_map<std::string, double>& b);

// Convenience wrapper: tokenises and computes TF-IDF for both strings,
// then returns their cosine similarity.
double textSimilarity(const std::string& a, const std::string& b);

} // namespace ranking
