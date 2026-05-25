#include "tfidf.hpp"
#include <algorithm>
#include <cctype>
#include <cmath>
#include <unordered_set>

namespace ranking {

static const std::unordered_set<std::string> STOP_WORDS = {
    "the","a","an","and","or","but","in","on","at","to","for","of","with",
    "is","are","was","were","be","been","being","have","has","had","do",
    "does","did","will","would","could","should","may","might","shall","can",
    "not","no","nor","so","yet","both","either","neither","than","as","if",
    "this","that","these","those","it","its","we","our","you","your","he",
    "she","they","their","my","his","her","i","me","him","us","them","by",
    "from","up","about","into","through","during","before","after","above",
    "below","between","each","more","also","such","when","which","who","how"
};

std::vector<std::string> tokenise(const std::string& text) {
    std::vector<std::string> tokens;
    std::string word;
    for (unsigned char c : text) {
        if (std::isalpha(c)) {
            word += static_cast<char>(std::tolower(c));
        } else if (!word.empty()) {
            if (word.size() > 2 && STOP_WORDS.find(word) == STOP_WORDS.end())
                tokens.push_back(word);
            word.clear();
        }
    }
    if (!word.empty() && word.size() > 2 && STOP_WORDS.find(word) == STOP_WORDS.end())
        tokens.push_back(word);
    return tokens;
}

std::unordered_map<std::string, double> computeTF(const std::string& text) {
    auto tokens = tokenise(text);
    if (tokens.empty()) return {};

    std::unordered_map<std::string, int> counts;
    for (const auto& t : tokens) counts[t]++;

    double total = static_cast<double>(tokens.size());
    std::unordered_map<std::string, double> tf;
    for (const auto& [term, cnt] : counts)
        tf[term] = cnt / total;
    return tf;
}

std::unordered_map<std::string, double> computeIDF(
    const std::vector<std::string>& corpus) {
    if (corpus.empty()) return {};

    int N = static_cast<int>(corpus.size());
    std::unordered_map<std::string, int> docFreq;
    for (const auto& doc : corpus) {
        auto tokens = tokenise(doc);
        std::unordered_set<std::string> seen(tokens.begin(), tokens.end());
        for (const auto& t : seen) docFreq[t]++;
    }

    std::unordered_map<std::string, double> idf;
    for (const auto& [term, df] : docFreq)
        idf[term] = std::log((N + 1.0) / (df + 1.0)) + 1.0;
    return idf;
}

std::unordered_map<std::string, double> computeTFIDF(
    const std::string& text,
    const std::unordered_map<std::string, double>& idf) {
    auto tf = computeTF(text);
    std::unordered_map<std::string, double> tfidf;
    for (const auto& [term, tfVal] : tf) {
        auto it = idf.find(term);
        tfidf[term] = tfVal * (it != idf.end() ? it->second : 1.0);
    }
    return tfidf;
}

std::unordered_map<std::string, double> computeTFIDFSingle(const std::string& text) {
    // IDF = 1.0 for all terms when there's only one document
    return computeTF(text);
}

std::vector<std::pair<std::string, double>> topKeywords(
    const std::unordered_map<std::string, double>& tfidf, int n) {
    std::vector<std::pair<std::string, double>> kws(tfidf.begin(), tfidf.end());
    int limit = std::min(n, static_cast<int>(kws.size()));
    std::partial_sort(kws.begin(), kws.begin() + limit, kws.end(),
        [](const auto& a, const auto& b) { return a.second > b.second; });
    kws.resize(limit);
    return kws;
}

} // namespace ranking
