# LRU Cache Sizing

This document explains the caching strategy used during git history analysis.

## Overview

When analyzing repository history, Clikd uses two LRU (Least Recently Used) caches to avoid redundant git operations:

1. **Commit Data Cache**: Stores processed commit information
2. **Tree Cache**: Stores git tree objects for diff computation

## Cache Configuration

Caches are configured in `.clikd.toml`:

```toml
[analysis]
commit_cache_size = 512  # default
tree_cache_size = 3      # default
```

## Commit Cache

**Default Size: 512 entries**

The commit cache stores processed commit data including:
- Affected projects (boolean hit buffer)
- Commit metadata
- Pre-computed diff results

### Why 512?

The default was chosen based on typical release cycles:
- Most releases contain 50-200 commits
- Cache size of 512 provides headroom for large releases
- Memory overhead per entry is minimal (~100-500 bytes)

For very large monorepos with thousands of commits between releases, increase this value:

```toml
[analysis]
commit_cache_size = 2048
```

## Tree Cache

**Default Size: 3 entries**

The tree cache stores git tree objects needed for diff computation.

### Why 3?

The minimal value of 3 is intentional:

```
Walking commit history linearly:

   Commit N-2    Commit N-1    Commit N (HEAD)
      │             │             │
      ▼             ▼             ▼
   Tree A        Tree B        Tree C
      ↑             ↑             ↑
      │             │             │
   [Evicted]     [Cached]      [Cached]
                    ↓             ↓
              diff(Tree B, Tree C) → compute which files changed
```

During linear history traversal:
1. Load current commit's tree (N)
2. Load parent commit's tree (N-1)
3. Compute diff between them
4. Move to parent (now N-1 becomes current, load new parent)

With cache size 3:
- Current tree
- Parent tree (for diff)
- One "spare" for merge commit handling

### Merge Commits

Merge commits have multiple parents, requiring additional tree lookups:

```
    Feature Branch
         │
    ┌────┴────┐
    │    B    │  ← second parent tree
    └────┬────┘
         │
    ┌────┴────┐     ┌────┴────┐
    │    M    │◄────│    A    │  ← first parent tree
    └────┬────┘     └─────────┘
         │
         ▼
      Current
```

The "spare" slot handles the second parent in most cases.

### When to Increase

Increase tree cache for repositories with:
- Frequent octopus merges (3+ parents)
- Complex branching strategies
- Parallel project walks (multiple revwalks active)

```toml
[analysis]
tree_cache_size = 8  # for complex merge patterns
```

## Memory Impact

| Cache Type | Entries | Approximate Memory |
|------------|---------|-------------------|
| Commit | 512 | 50-250 KB |
| Tree | 3 | 1-50 KB per tree |

Tree objects can vary significantly based on directory structure.

## Algorithm Details

The analysis algorithm walks commits for each project:

```rust
for proj_idx in 0..projects.len() {
    let mut walk = repo.revwalk()?;
    walk.push_head()?;

    // Hide commits before last release tag
    if let Some(tag_info) = &histories[proj_idx].release_tag {
        walk.hide(tag_info.commit)?;
    }

    for oid in walk {
        // Check commit cache first
        if !commit_data.contains(&oid) {
            // Cache miss: load trees, compute diff
            let cur_tree = trees.pop_or_load(&ctid)?;
            let parent_tree = trees.pop_or_load(&ptid)?;
            let diff = repo.diff_tree_to_tree(...)?;

            // Store back in cache
            trees.put(ctid, cur_tree);
            trees.put(ptid, parent_tree);

            // Process diff, update hit buffer
            commit_data.put(oid, processed_data);
        }

        // Use cached data
        histories[proj_idx].add_commit(commit_data.get(&oid)?);
    }
}
```

Key insight: When analyzing multiple projects, commits are often visited multiple times. The commit cache prevents re-computing diffs for the same commits.

## Performance Tuning

### Symptom: Slow analysis with many projects

If `clikd release status` is slow:

1. Check if commits are being re-processed:
   ```bash
   RUST_LOG=debug clikd release status 2>&1 | grep "cache miss"
   ```

2. Increase commit cache:
   ```toml
   [analysis]
   commit_cache_size = 2048
   ```

### Symptom: High memory usage

If memory is a concern:

1. Reduce commit cache for small release cycles:
   ```toml
   [analysis]
   commit_cache_size = 128
   ```

2. Tree cache should remain at least 3

### Benchmarking

Measure analysis performance:

```bash
time clikd release status --format=json > /dev/null
```

Compare with different cache sizes to find optimal values for your repository.

## Trade-offs

| Setting | Small Cache | Large Cache |
|---------|-------------|-------------|
| Memory | Lower | Higher |
| Speed (first run) | Same | Same |
| Speed (multi-project) | Slower | Faster |
| Best for | Simple repos | Large monorepos |

## Recommendations

| Repository Type | commit_cache_size | tree_cache_size |
|-----------------|-------------------|-----------------|
| Single project | 128 | 3 |
| Small monorepo (2-5 projects) | 256 | 3 |
| Medium monorepo (5-20 projects) | 512 | 3 |
| Large monorepo (20+ projects) | 1024-2048 | 5-8 |
| Complex merge strategy | 512 | 8-16 |
