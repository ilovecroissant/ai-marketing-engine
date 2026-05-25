#pragma once
#include <string>
#include <unordered_map>
#include <utility>
#include <vector>

namespace ranking {

// Split text into lowercase tokens, stripping punctuation and stop words.
std::vector<std::string> tokenise(const std::string& text);

// Term frequency for a single document: count(t) / total_words.
std::unordered_map<std::string, double> computeTF(const std::string& text);

// Inverse document frequency across a corpus of documents.
// Uses smoothed IDF: log((N+1)/(df+1)) + 1 to avoid division-by-zero.
std::unordered_map<std::string, double> computeIDF(
    const std::vector<std::string>& corpus);

// TF-IDF vector for one document given a pre-computed IDF table.
std::unordered_map<std::string, double> computeTFIDF(
    const std::string& text,
    const std::unordered_map<std::string, double>& idf);

// TF-IDF for a single document treated as its own 1-doc corpus (IDF=1).
// Used when no external corpus is available.
std::unordered_map<std::string, double> computeTFIDFSingle(const std::string& text);

// Return top-N keywords sorted by descending TF-IDF score.
std::vector<std::pair<std::string, double>> topKeywords(
    const std::unordered_map<std::string, double>& tfidf, int n = 10);

} // namespace ranking
