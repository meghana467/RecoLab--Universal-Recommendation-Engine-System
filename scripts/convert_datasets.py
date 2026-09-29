#!/usr/bin/env python3
"""
Convert Retailrocket e-commerce + MIND news datasets into Gorse-compatible
JSON files that replace the synthetic demo data.

Retailrocket mappings (per colleague):
  visitorid       -> UserId
  itemid          -> ItemId  (prefixed "rr_")
  view            -> FeedbackType "view"
  addtocart       -> FeedbackType "cart"
  transaction     -> FeedbackType "purchase"
  categoryid prop -> Item category
  other props     -> Item labels

MIND mappings (per colleague):
  UserID          -> UserId  (prefixed "mn_")
  NewsID          -> ItemId  (prefixed "mn_")
  clicked (-1)    -> FeedbackType "click"
  unclicked (-0)  -> FeedbackType "read"
  history clicks  -> FeedbackType "click"
  Category        -> Item category
  SubCategory     -> Item label
  Title entities  -> Item labels
"""

import csv
import json
import os
import datetime

# ── paths ──────────────────────────────────────────────────────────────────
RR_DIR   = "/Users/jesvin/Downloads/archive-2"
MIND_DIR = "/Users/jesvin/Downloads/archive/MINDsmall_train"
OUT_DIR  = "/Users/jesvin/Downloads/meg/recolab-gorse-demo/data"

# Demo limits – enough for a convincing demo, small enough to load quickly
RR_MAX_USERS    = 2000
RR_MAX_FEEDBACK = 50000
MIND_MAX_USERS  = 2000
MIND_MAX_FEEDBACK = 50000

# ── helpers ────────────────────────────────────────────────────────────────
def ts_from_ms(ms_str):
    """Retailrocket timestamps are Unix milliseconds."""
    try:
        return datetime.datetime.utcfromtimestamp(int(ms_str) / 1000).strftime("%Y-%m-%dT%H:%M:%SZ")
    except Exception:
        return "2019-01-01T00:00:00Z"

def ts_from_mind(dt_str):
    """MIND timestamps like '11/11/2019 9:05:58 AM'."""
    try:
        dt = datetime.datetime.strptime(dt_str, "%m/%d/%Y %I:%M:%S %p")
        return dt.strftime("%Y-%m-%dT%H:%M:%SZ")
    except Exception:
        return "2019-01-01T00:00:00Z"

# ══════════════════════════════════════════════════════════════════════════
# 1. RETAILROCKET
# ══════════════════════════════════════════════════════════════════════════
print("=== Processing Retailrocket ===")

EVENT_TYPE_MAP = {"view": "view", "addtocart": "cart", "transaction": "purchase"}

# -- Pass 1: collect top users by event count
print("  Scanning events for top users…")
user_counts = {}
with open(os.path.join(RR_DIR, "events.csv"), newline="") as f:
    reader = csv.DictReader(f)
    for row in reader:
        uid = row["visitorid"]
        user_counts[uid] = user_counts.get(uid, 0) + 1

top_users = set(sorted(user_counts, key=user_counts.get, reverse=True)[:RR_MAX_USERS])
print(f"  Selected {len(top_users)} Retailrocket users")

# -- Pass 2: collect feedback + item ids for those users
print("  Collecting feedback…")
rr_feedback = []
rr_item_ids = set()
with open(os.path.join(RR_DIR, "events.csv"), newline="") as f:
    reader = csv.DictReader(f)
    for row in reader:
        uid = row["visitorid"]
        if uid not in top_users:
            continue
        ft = EVENT_TYPE_MAP.get(row["event"])
        if ft is None:
            continue
        iid = "rr_" + row["itemid"]
        rr_item_ids.add(iid)
        rr_feedback.append({
            "FeedbackType": ft,
            "UserId": "rr_" + uid,
            "ItemId": iid,
            "Timestamp": ts_from_ms(row["timestamp"]),
        })
        if len(rr_feedback) >= RR_MAX_FEEDBACK:
            break

print(f"  Collected {len(rr_feedback)} feedback, {len(rr_item_ids)} items")

# -- Build item properties: categoryid -> category, others -> labels
print("  Building item properties from item_properties files…")
item_category = {}   # itemid -> category string
item_labels   = {}   # itemid -> set of labels

def process_props_file(path):
    with open(path, newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            iid = "rr_" + row["itemid"]
            if iid not in rr_item_ids:
                continue
            prop  = row["property"]
            value = row["value"].strip()
            if not value or value == "n/a":
                continue
            if prop == "categoryid":
                item_category[iid] = "cat_" + value
            else:
                # value may be space-separated token list
                tokens = [v for v in value.split() if not v.startswith("n")]
                if tokens:
                    item_labels.setdefault(iid, set()).update(tokens[:5])

process_props_file(os.path.join(RR_DIR, "item_properties_part1.csv"))
process_props_file(os.path.join(RR_DIR, "item_properties_part2.csv"))

# -- Build Gorse items
rr_items = []
for iid in rr_item_ids:
    cat  = item_category.get(iid, "uncategorized")
    lbls = list(item_labels.get(iid, set()))[:10]
    rr_items.append({
        "ItemId":     iid,
        "IsHidden":   False,
        "Categories": ["products", cat],
        "Labels":     lbls,
        "Comment":    f"Retailrocket item {iid}",
    })

# -- Build Gorse users
rr_users = [{"UserId": "rr_" + u, "Labels": [], "Comment": "Retailrocket visitor"} for u in top_users]

print(f"  Retailrocket: {len(rr_users)} users, {len(rr_items)} items, {len(rr_feedback)} feedback")

# ══════════════════════════════════════════════════════════════════════════
# 2. MIND news
# ══════════════════════════════════════════════════════════════════════════
print("\n=== Processing MIND ===")

# -- Parse news.tsv  (tab-separated, no header)
# columns: NewsID  Category  SubCategory  Title  Abstract  URL  TitleEntities  AbstractEntities
print("  Parsing news.tsv…")
mind_items = {}
with open(os.path.join(MIND_DIR, "news.tsv"), newline="", encoding="utf-8") as f:
    reader = csv.reader(f, delimiter="\t")
    for row in reader:
        if len(row) < 7:
            continue
        news_id, category, subcategory, title, abstract, url, title_entities = row[:7]
        labels = [subcategory] if subcategory else []
        try:
            ents = json.loads(title_entities)
            labels += [e["Label"] for e in ents if "Label" in e][:5]
        except Exception:
            pass
        mind_items["mn_" + news_id] = {
            "ItemId":     "mn_" + news_id,
            "IsHidden":   False,
            "Categories": ["news", category],
            "Labels":     list(dict.fromkeys(labels))[:10],
            "Comment":    title[:120],
        }

print(f"  Parsed {len(mind_items)} news items")

# -- Parse behaviors.tsv (tab-separated, no header)
# columns: ImpressionID  UserID  Time  History  Impressions
print("  Parsing behaviors.tsv…")
mind_users_seen = {}  # uid -> count
with open(os.path.join(MIND_DIR, "behaviors.tsv"), newline="", encoding="utf-8") as f:
    reader = csv.reader(f, delimiter="\t")
    for row in reader:
        if len(row) < 2:
            continue
        uid = row[1]
        mind_users_seen[uid] = mind_users_seen.get(uid, 0) + 1

top_mind_users = set(sorted(mind_users_seen, key=mind_users_seen.get, reverse=True)[:MIND_MAX_USERS])
print(f"  Selected {len(top_mind_users)} MIND users")

mind_feedback = []
mind_item_ids_used = set()

with open(os.path.join(MIND_DIR, "behaviors.tsv"), newline="", encoding="utf-8") as f:
    reader = csv.reader(f, delimiter="\t")
    for row in reader:
        if len(row) < 5:
            continue
        _, uid, time_str, history, impressions = row[0], row[1], row[2], row[3], row[4]
        if uid not in top_mind_users:
            continue
        ts = ts_from_mind(time_str)
        gorse_uid = "mn_" + uid

        # History: previously clicked articles → "click"
        if history:
            for nid in history.strip().split():
                item_key = "mn_" + nid
                if item_key in mind_items:
                    mind_item_ids_used.add(item_key)
                    mind_feedback.append({
                        "FeedbackType": "click",
                        "UserId":       gorse_uid,
                        "ItemId":       item_key,
                        "Timestamp":    ts,
                    })

        # Impressions: N12345-1 (clicked) or N12345-0 (not clicked)
        for imp in impressions.strip().split():
            parts = imp.rsplit("-", 1)
            if len(parts) != 2:
                continue
            nid, label = parts
            item_key = "mn_" + nid
            ft = "click" if label == "1" else "read"
            if item_key in mind_items:
                mind_item_ids_used.add(item_key)
                mind_feedback.append({
                    "FeedbackType": ft,
                    "UserId":       gorse_uid,
                    "ItemId":       item_key,
                    "Timestamp":    ts,
                })

        if len(mind_feedback) >= MIND_MAX_FEEDBACK:
            break

mind_users = [{"UserId": "mn_" + u, "Labels": [], "Comment": "MIND news reader"} for u in top_mind_users]
mind_items_list = [v for k, v in mind_items.items() if k in mind_item_ids_used]

print(f"  MIND: {len(mind_users)} users, {len(mind_items_list)} items, {len(mind_feedback)} feedback")

# ══════════════════════════════════════════════════════════════════════════
# 3. Merge and write
# ══════════════════════════════════════════════════════════════════════════
print("\n=== Writing output files ===")

all_users    = rr_users + mind_users
all_feedback = rr_feedback + mind_feedback

# Keep existing expo_booths.json (synthetic, no real dataset equivalent)
# Write product_items.json from Retailrocket
# Write news_items.json from MIND

def write_json(path, data):
    with open(path, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    print(f"  Wrote {len(data):,} records → {path}")

write_json(os.path.join(OUT_DIR, "users.json"),         all_users)
write_json(os.path.join(OUT_DIR, "product_items.json"), rr_items)
write_json(os.path.join(OUT_DIR, "news_items.json"),    mind_items_list)
write_json(os.path.join(OUT_DIR, "feedback.json"),      all_feedback)

print(f"\nDone. Totals: {len(all_users):,} users | "
      f"{len(rr_items)+len(mind_items_list):,} items | "
      f"{len(all_feedback):,} feedback events")
