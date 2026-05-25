cmake_minimum_required(VERSION 3.16)
project(ranking-engine CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# MSVC: treat source as UTF-8, enable all warnings
if(MSVC)
    add_compile_options(/utf-8 /W3)
endif()

include(FetchContent)

# cpp-httplib — header-only HTTP server/client
FetchContent_Declare(
    httplib
    GIT_REPOSITORY https://github.com/yhirose/cpp-httplib.git
    GIT_TAG        v0.18.3
    GIT_SHALLOW    TRUE
)
FetchContent_MakeAvailable(httplib)

# nlohmann/json — header-only JSON
FetchContent_Declare(
    nlohmann_json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG        v3.11.3
    GIT_SHALLOW    TRUE
)
FetchContent_MakeAvailable(nlohmann_json)

# Core library shared between server and tests
add_library(ranking_core
    src/tfidf.cpp
    src/cosine.cpp
    src/ranking.cpp
)
target_include_directories(ranking_core PUBLIC src)
target_link_libraries(ranking_core PUBLIC nlohmann_json::nlohmann_json)

# HTTP server executable
add_executable(ranking-server src/main.cpp)
target_link_libraries(ranking-server PRIVATE ranking_core httplib::httplib)

# Unit tests — run with: ctest --output-on-failure
enable_testing()
add_executable(ranking-tests tests/test_main.cpp)
target_link_libraries(ranking-tests PRIVATE ranking_core)
add_test(NAME unit_tests COMMAND ranking-tests)
