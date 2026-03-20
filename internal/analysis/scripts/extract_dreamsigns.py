#!/usr/bin/env python3
"""
Extract dreamsigns from dream journal entries.

Dreamsigns are recurring elements in dreams that can help identify
lucid dreaming triggers. Uses NER, noun chunks, and keyphrase extraction.
"""

import argparse
import json
import sqlite3
import sys
from collections import Counter
from pathlib import Path

import numpy as np
from keybert import KeyBERT
from sklearn.cluster import KMeans
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics import silhouette_score


def fetch_dreams(db_path: str) -> list[dict]:
    """Fetch all dreams from SQLite database."""
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    cursor = conn.cursor()

    cursor.execute("""
        SELECT id, content, created_at
        FROM dreams
        ORDER BY created_at DESC
    """)

    dreams = [dict(row) for row in cursor.fetchall()]
    conn.close()

    return dreams


def preprocess_text(text: str) -> str:
    """Basic text preprocessing."""
    return text.lower().strip()


def load_spacy_model():
    """Lazy load spaCy model."""
    import spacy

    try:
        return spacy.load("en_core_web_sm")
    except OSError:
        print("Downloading spaCy model...", file=sys.stderr)
        spacy.cli.download("en_core_web_sm")
        return spacy.load("en_core_web_sm")


def extract_entities(text: str, nlp) -> list[str]:
    """Extract named entities (people, places, organizations)."""
    doc = nlp(text)
    entities = []

    for ent in doc.ents:
        if ent.label_ in (
            "PERSON",
            "ORG",
            "GPE",
            "LOC",
            "PRODUCT",
            "EVENT",
            "WORK_OF_ART",
        ):
            entities.append(ent.text.lower())

    return entities


def extract_noun_chunks(text: str, nlp) -> list[str]:
    """Extract noun phrases (descriptive objects/concepts)."""
    doc = nlp(text)
    chunks = []

    for chunk in doc.noun_chunks:
        text = chunk.text.lower().strip()
        words = text.split()

        if len(words) > 3 or len(text) <= 2:
            continue

        if text in STOP_WORDS or any(w in STOP_WORDS for w in words):
            continue

        root = chunk.root
        if root.pos_ in ("PRON", "DET"):
            continue

        chunks.append(text)

    return chunks


def extract_keyphrases(texts: list[str], top_n: int = 50) -> list[str]:
    """Extract keyphrases using KeyBERT."""
    if not texts or len(texts) < 2:
        return []

    kw_model = KeyBERT()
    all_phrases = []

    for text in texts:
        keywords = kw_model.extract_keywords(
            text, keyphrase_ngram_range=(1, 2), stop_words="english", top_n=5
        )
        all_phrases.extend([kw for kw, _ in keywords])

    return all_phrases


# Generic verbs and words to filter out
STOP_WORDS = {
    "get",
    "got",
    "getting",
    "keep",
    "kept",
    "keeping",
    "feel",
    "felt",
    "see",
    "saw",
    "seen",
    "looking",
    "look",
    "looked",
    "go",
    "went",
    "going",
    "gone",
    "come",
    "came",
    "coming",
    "take",
    "took",
    "taken",
    "taking",
    "make",
    "made",
    "making",
    "think",
    "thought",
    "thinking",
    "know",
    "knew",
    "knowing",
    "say",
    "said",
    "saying",
    "tell",
    "told",
    "start",
    "started",
    "try",
    "tried",
    "trying",
    "want",
    "wanted",
    "need",
    "needed",
    "seem",
    "seemed",
    "use",
    "used",
    "put",
    "let",
    "like",
    "love",
    "hate",
    "fear",
    "worry",
    "seem",
    "seemed",
    "help",
    "helped",
    "helping",
    "ask",
    "asked",
    "asking",
    "show",
    "showed",
    "giving",
    "give",
    "gave",
    "given",
    "leave",
    "left",
    "leaving",
    "find",
    "found",
    "finding",
    "right",
    "wrong",
    "good",
    "bad",
    "old",
    "new",
    "big",
    "small",
    "long",
    "short",
    "high",
    "low",
    # Pronouns and generic references
    "i",
    "me",
    "my",
    "mine",
    "myself",
    "you",
    "your",
    "yours",
    "yourself",
    "he",
    "him",
    "his",
    "himself",
    "she",
    "her",
    "hers",
    "herself",
    "it",
    "its",
    "itself",
    "we",
    "us",
    "our",
    "ours",
    "ourselves",
    "they",
    "them",
    "their",
    "theirs",
    "themselves",
    "themself",
    "this",
    "that",
    "these",
    "those",
    "here",
    "there",
    "everywhere",
    "something",
    "someone",
    "somebody",
    "anything",
    "anyone",
    "anybody",
    "everything",
    "everyone",
    "everybody",
    "nothing",
    "no one",
    "nobody",
    "one",
    "ones",
    "other",
    "others",
    "another",
    "each",
    "every",
    # Generic time/place references
    "time",
    "day",
    "night",
    "morning",
    "evening",
    "today",
    "yesterday",
    "tomorrow",
    "week",
    "month",
    "year",
    "moment",
    "second",
    "minute",
    "hour",
    "place",
    "home",
    "house",
    "room",
    "way",
    "part",
    "thing",
    "things",
    "something",
    "everything",
    "nothing",
    "dream",
    "dreams",
}


def is_valid_dreamsign(term: str) -> bool:
    """Check if a term is a valid dreamsign (not a generic verb)."""
    term_lower = term.lower().strip()
    words = term_lower.split()

    if term_lower in STOP_WORDS:
        return False

    if all(w in STOP_WORDS for w in words):
        return False

    if len(term_lower) < 3:
        return False

    return True


def extract_dreamsigns_from_texts(
    texts: list[str], nlp, min_freq: int = 2, top_n: int = 20
) -> list[tuple[str, int]]:
    """Extract dreamsigns from texts using multiple methods."""

    all_candidates = []

    for text in texts:
        all_candidates.extend(extract_entities(text, nlp))
        all_candidates.extend(extract_noun_chunks(text, nlp))

    all_candidates.extend(extract_keyphrases(texts, top_n=50))

    filtered = [c for c in all_candidates if is_valid_dreamsign(c)]

    counter = Counter(filtered)

    return [
        (term, count) for term, count in counter.most_common(top_n) if count >= min_freq
    ]


def cluster_dreams(
    dreams: list[dict], n_clusters: int | None = None, min_dreams: int = 5
) -> dict:
    """
    Cluster dreams and extract meaningful dreamsigns per cluster.

    Args:
        dreams: List of dream dictionaries
        n_clusters: Number of clusters (auto-determined if None)
        min_dreams: Minimum dreams required for analysis

    Returns:
        Dictionary with clusters and their dreamsigns
    """
    if len(dreams) < min_dreams:
        return {
            "error": f"Need at least {min_dreams} dreams, found {len(dreams)}",
            "dream_count": len(dreams),
            "clusters": [],
        }

    contents = [preprocess_text(d["content"]) for d in dreams]

    # TF-IDF for clustering
    vectorizer = TfidfVectorizer(
        max_features=1000,
        stop_words="english",
        ngram_range=(1, 2),
        min_df=2,
    )

    try:
        tfidf_matrix = vectorizer.fit_transform(contents)
    except ValueError as e:
        return {
            "error": f"Failed to vectorize dreams: {str(e)}",
            "dream_count": len(dreams),
            "clusters": [],
        }

    # Auto-determine clusters
    if n_clusters is None:
        max_clusters = min(10, len(dreams) // 2)
        if max_clusters < 2:
            n_clusters = 2
        else:
            best_score, best_k = -1, 2
            for k in range(2, max_clusters + 1):
                kmeans = KMeans(n_clusters=k, random_state=42, n_init="auto")
                labels = kmeans.fit_predict(tfidf_matrix)
                score = silhouette_score(tfidf_matrix, labels)
                if score > best_score:
                    best_score, best_k = score, k
            n_clusters = best_k

    kmeans = KMeans(n_clusters=n_clusters, random_state=42, n_init="auto")
    labels = kmeans.fit_predict(tfidf_matrix)

    # Load spaCy for entity extraction
    nlp = load_spacy_model()

    # Extract dreamsigns per cluster
    clusters = []
    for i in range(n_clusters):
        cluster_indices = [idx for idx, label in enumerate(labels) if label == i]
        cluster_contents = [dreams[idx]["content"] for idx in cluster_indices]

        dreamsigns = extract_dreamsigns_from_texts(
            cluster_contents, nlp, min_freq=1, top_n=5
        )

        clusters.append(
            {
                "cluster_id": i,
                "dream_count": len(cluster_indices),
                "top_terms": [term for term, count in dreamsigns],
                "dream_ids": [dreams[idx]["id"] for idx in cluster_indices],
            }
        )

    return {
        "dream_count": len(dreams),
        "n_clusters": n_clusters,
        "clusters": clusters,
    }


def main():
    parser = argparse.ArgumentParser(
        description="Extract dreamsigns from dream journal"
    )
    parser.add_argument(
        "--db-path", type=str, default="dreams.db", help="Path to SQLite database"
    )
    parser.add_argument("--output", type=str, help="Output file path (JSON)")
    parser.add_argument(
        "--n-clusters", type=int, help="Number of clusters (auto if not set)"
    )
    parser.add_argument(
        "--min-dreams", type=int, default=5, help="Minimum dreams required"
    )

    args = parser.parse_args()

    db_path = Path(args.db_path)
    if not db_path.exists():
        print(f"Error: Database not found at {args.db_path}", file=sys.stderr)
        sys.exit(1)

    dreams = fetch_dreams(args.db_path)
    result = cluster_dreams(
        dreams, n_clusters=args.n_clusters, min_dreams=args.min_dreams
    )

    json_output = json.dumps(result, indent=2)

    if args.output:
        output_path = Path(args.output)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        output_path.write_text(json_output)
        print(f"Results saved to {args.output}")
    else:
        print(json_output)


if __name__ == "__main__":
    main()
