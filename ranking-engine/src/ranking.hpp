#pragma once
#include <string>
#include <vector>

namespace ranking {

struct ContentScore {
    double seo;          // keyword coverage and density [0, 1]
    double engagement;   // content length and sentence variety [0, 1]
    double readability;  // word length and syllable complexity [0, 1]
    double total;        // 0.4*seo + 0.3*engagement + 0.3*readability
};

// Score a piece of content against a set of target keywords.
// All sub-scores are normalised to [0, 1].
ContentScore scoreContent(
    const std::string& text,
    const std::vector<std::string>& targetKeywords);

} // namespace ranking
