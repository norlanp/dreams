#!/usr/bin/env python3
"""
Extract dreamsigns from dream journal entries using clustering.

Dreamsigns are recurring elements in dreams that can help identify
lucid dreaming triggers. This script uses TF-IDF + K-means clustering
to automatically identify common themes.
"""

import argparse
import json
import sqlite3
import sys
from pathlib import Path

import numpy as np
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


def extract_dreamsigns(
    dreams: list[dict], n_clusters: int | None = None, min_dreams: int = 5
) -> dict:
    """
    Extract dreamsigns using TF-IDF + K-means clustering.

    Args:
        dreams: List of dream dictionaries with 'content' key
        n_clusters: Number of clusters (auto-determined if None)
        min_dreams: Minimum number of dreams required for analysis

    Returns:
        Dictionary containing clusters and their representative terms
    """
    if len(dreams) < min_dreams:
        return {
            "error": f"Need at least {min_dreams} dreams for clustering, found {len(dreams)}",
            "dream_count": len(dreams),
            "clusters": [],
        }

    contents = [preprocess_text(d["content"]) for d in dreams]

    # TF-IDF vectorization
    vectorizer = TfidfVectorizer(
        max_features=1000,
        stop_words="english",
        ngram_range=(1, 2),  # Unigrams and bigrams
        min_df=2,  # Term must appear in at least 2 dreams
    )

    try:
        tfidf_matrix = vectorizer.fit_transform(contents)
    except ValueError as e:
        return {
            "error": f"Failed to vectorize dreams: {str(e)}",
            "dream_count": len(dreams),
            "clusters": [],
        }

    # Auto-determine optimal number of clusters
    if n_clusters is None:
        max_clusters = min(10, len(dreams) // 2)
        if max_clusters < 2:
            n_clusters = 2
        else:
            best_score = -1
            best_k = 2

            for k in range(2, max_clusters + 1):
                kmeans = KMeans(n_clusters=k, random_state=42, n_init=10)
                labels = kmeans.fit_predict(tfidf_matrix)
                score = silhouette_score(tfidf_matrix, labels)

                if score > best_score:
                    best_score = score
                    best_k = k

            n_clusters = best_k

    # Final clustering
    kmeans = KMeans(n_clusters=n_clusters, random_state=42, n_init=10)
    labels = kmeans.fit_predict(tfidf_matrix)

    # Get feature names
    feature_names = vectorizer.get_feature_names_out()

    # Extract top terms for each cluster
    clusters = []
    for i in range(n_clusters):
        center = kmeans.cluster_centers_[i]
        top_indices = center.argsort()[-5:][::-1]
        top_terms = [feature_names[idx] for idx in top_indices]

        # Count dreams in this cluster
        dream_indices = [idx for idx, label in enumerate(labels) if label == i]

        clusters.append(
            {
                "cluster_id": i,
                "dream_count": len(dream_indices),
                "top_terms": top_terms,
                "dream_ids": [dreams[idx]["id"] for idx in dream_indices],
            }
        )

    return {"dream_count": len(dreams), "n_clusters": n_clusters, "clusters": clusters}


def main():
    parser = argparse.ArgumentParser(
        description="Extract dreamsigns from dream journal using clustering"
    )
    parser.add_argument(
        "--db-path", type=str, default="var/dreams.db", help="Path to SQLite database"
    )
    parser.add_argument("--output", type=str, help="Output file path (JSON format)")
    parser.add_argument(
        "--n-clusters",
        type=int,
        help="Number of clusters (auto-determined if not specified)",
    )
    parser.add_argument(
        "--min-dreams",
        type=int,
        default=5,
        help="Minimum number of dreams required for analysis",
    )

    args = parser.parse_args()

    # Check if database exists
    db_path = Path(args.db_path)
    if not db_path.exists():
        print(f"Error: Database not found at {args.db_path}", file=sys.stderr)
        sys.exit(1)

    # Fetch and analyze dreams
    dreams = fetch_dreams(args.db_path)
    result = extract_dreamsigns(
        dreams, n_clusters=args.n_clusters, min_dreams=args.min_dreams
    )

    # Output results
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
