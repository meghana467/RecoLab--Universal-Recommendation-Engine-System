#!/usr/bin/env python3
"""Generate all demo data JSON files for RecoLab."""

import json
import random
import os
from datetime import datetime, timedelta

random.seed(42)

BASE_DIR = os.path.join(os.path.dirname(__file__), "..", "data")
os.makedirs(BASE_DIR, exist_ok=True)

# ── helpers ─────────────────────────────────────────────────────────────────

def rand_ts(start_day=1, end_day=23):
    d = random.randint(start_day, end_day)
    h = random.randint(0, 23)
    m = random.randint(0, 59)
    return f"2026-06-{d:02d}T{h:02d}:{m:02d}:00Z"

def pick(lst, k=None):
    if k is None:
        return random.choice(lst)
    return random.sample(lst, min(k, len(lst)))

# ── USERS ───────────────────────────────────────────────────────────────────

tech_interests   = ["ai", "technology", "startup", "india", "ml", "cloud", "semiconductor", "software"]
trade_interests  = ["trade", "export", "manufacturing", "b2b", "supply-chain", "logistics", "global-trade"]
sports_interests = ["sports", "cricket", "football", "olympics", "tennis", "fitness", "ipl"]
biz_interests    = ["business", "finance", "economy", "markets", "investment", "banking"]
life_interests   = ["lifestyle", "health", "entertainment", "food", "travel", "wellness"]
buyer_interests  = ["import", "procurement", "sourcing", "b2b", "trade", "expo"]

users = []

for i in range(1, 31):
    uid = f"u_{i:03d}"
    labels = pick(tech_interests, random.randint(3, 6))
    users.append({"UserId": uid, "Labels": labels, "Comment": f"Tech enthusiast #{i}"})

for i in range(31, 51):
    uid = f"u_{i:03d}"
    labels = pick(trade_interests, random.randint(3, 5))
    users.append({"UserId": uid, "Labels": labels, "Comment": f"Trade professional #{i}"})

for i in range(51, 66):
    uid = f"u_{i:03d}"
    labels = pick(sports_interests, random.randint(3, 5))
    users.append({"UserId": uid, "Labels": labels, "Comment": f"Sports fan #{i}"})

for i in range(66, 81):
    uid = f"u_{i:03d}"
    labels = pick(biz_interests, random.randint(3, 5))
    users.append({"UserId": uid, "Labels": labels, "Comment": f"Business reader #{i}"})

for i in range(81, 91):
    uid = f"u_{i:03d}"
    labels = pick(life_interests, random.randint(2, 4))
    users.append({"UserId": uid, "Labels": labels, "Comment": f"Lifestyle reader #{i}"})

for i in range(91, 101):
    uid = f"u_{i:03d}"
    labels = pick(buyer_interests, random.randint(3, 5))
    users.append({"UserId": uid, "Labels": labels, "Comment": f"Buyer #{i}"})

# Named users
users += [
    {"UserId": "u_ai_reader",      "Labels": ["ai", "technology", "startup", "ml", "india"], "Comment": "AI-focused reader"},
    {"UserId": "u_trade_reader",   "Labels": ["trade", "export", "b2b", "manufacturing"],    "Comment": "Trade & export reader"},
    {"UserId": "u_sports_reader",  "Labels": ["sports", "cricket", "football", "olympics"],   "Comment": "Sports enthusiast"},
    {"UserId": "buyer_energy_001", "Labels": ["solar", "renewable", "energy", "procurement", "import"], "Comment": "Energy sector buyer"},
    {"UserId": "buyer_textile_001","Labels": ["textile", "fashion", "procurement", "import"], "Comment": "Textile buyer"},
    {"UserId": "buyer_energy_009", "Labels": ["solar", "wind", "energy", "b2b"],              "Comment": "Renewable energy buyer"},
    {"UserId": "buyer_import_014", "Labels": ["import", "chemicals", "procurement"],           "Comment": "Chemicals importer"},
]

with open(os.path.join(BASE_DIR, "users.json"), "w") as f:
    json.dump(users, f, indent=2)
print(f"Generated {len(users)} users")

# ── NEWS ITEMS ───────────────────────────────────────────────────────────────

news_templates = [
    # AI/Tech
    ("Indian AI semiconductor startup raises $120M Series B", ["ai", "startup", "india", "semiconductor", "technology"]),
    ("OpenAI releases GPT-5 with improved reasoning capabilities", ["ai", "technology", "ml", "openai"]),
    ("India emerges as top AI talent destination in Asia", ["ai", "india", "technology", "startup"]),
    ("Cloud giants race to deploy quantum computing infrastructure", ["cloud", "technology", "quantum"]),
    ("Deep learning breakthrough reduces model training costs 80%", ["ai", "ml", "technology"]),
    ("Bengaluru startup ecosystem attracts $2B in Q2 2026", ["startup", "india", "technology", "investment"]),
    ("AI-powered drug discovery firm IPOs on NSE", ["ai", "biotech", "india", "startup"]),
    ("Government launches $5B national AI mission", ["ai", "india", "policy", "technology"]),
    ("Meta unveils AR glasses with AI assistant built-in", ["ai", "technology", "hardware", "meta"]),
    ("India's semiconductor fab receives production approval", ["semiconductor", "india", "technology", "manufacturing"]),
    ("New large language model beats GPT-4 on coding benchmarks", ["ai", "ml", "technology", "software"]),
    ("Autonomous vehicle startup partners with major automaker", ["ai", "technology", "automotive", "startup"]),
    ("AI regulation framework proposed by Indian Parliament", ["ai", "india", "policy", "technology"]),
    ("Robotics firm deploys 10,000 warehouse bots across India", ["robotics", "ai", "india", "technology"]),
    ("Chip design startup from IIT raises pre-IPO funding", ["semiconductor", "startup", "india", "technology"]),
    # Trade
    ("India-EU free trade agreement enters final negotiation phase", ["trade", "india", "global-trade", "policy"]),
    ("Indian textile exports hit record $18B in FY2026", ["trade", "textile", "india", "export"]),
    ("New logistics corridor cuts India-Middle East transit time", ["logistics", "trade", "india", "supply-chain"]),
    ("Export credit scheme extended for MSME manufacturers", ["trade", "export", "manufacturing", "india"]),
    ("India surpasses China as top garment exporter to EU", ["trade", "textile", "india", "export"]),
    ("Chemical industry sees 35% export growth in H1 2026", ["chemicals", "trade", "export", "india"]),
    ("India-Africa trade summit announces $10B investment pledge", ["trade", "india", "global-trade", "africa"]),
    ("Solar panel export opportunity widens as US tariff eases", ["solar", "renewable", "export", "trade"]),
    ("Global supply chain realignment benefits Indian manufacturers", ["manufacturing", "trade", "supply-chain", "india"]),
    ("B2B e-commerce platform for exports raises Series C", ["b2b", "trade", "startup", "export"]),
    # Sports
    ("India wins Test series against Australia 3-1", ["sports", "cricket", "india"]),
    ("IPL 2026 final draws record 100M viewers", ["sports", "cricket", "ipl", "india"]),
    ("Neeraj Chopra defends javelin world title in Paris", ["sports", "athletics", "india", "olympics"]),
    ("India qualifies for 2026 FIFA World Cup", ["sports", "football", "india"]),
    ("PV Sindhu announces comeback after injury recovery", ["sports", "badminton", "india"]),
    ("Indian Premier League expansion adds two new teams", ["sports", "cricket", "ipl", "india"]),
    ("Chess Olympiad 2026: India wins gold in open category", ["sports", "chess", "india"]),
    ("Rohit Sharma breaks Sachin's ODI century record", ["sports", "cricket", "india"]),
    ("Pro Kabaddi League attracts global broadcast deal", ["sports", "kabaddi", "india"]),
    ("Indian hockey team wins Asian Games gold", ["sports", "hockey", "india", "olympics"]),
    # Business/Finance
    ("RBI holds repo rate at 5.75% amid stable inflation", ["finance", "economy", "india", "banking"]),
    ("Sensex hits 100,000 milestone for first time", ["markets", "finance", "india", "economy"]),
    ("Foreign direct investment in India reaches $85B in 2026", ["investment", "india", "economy", "business"]),
    ("India GDP growth forecast raised to 7.8% for FY26", ["economy", "india", "business"]),
    ("Startup unicorn count reaches 200 in India", ["startup", "india", "business", "investment"]),
    ("Digital payments in India cross $500B monthly volume", ["finance", "india", "technology", "economy"]),
    ("Insurance sector opens to 100% FDI", ["finance", "india", "investment", "policy"]),
    ("Green bond market grows 120% year-on-year", ["finance", "environment", "investment", "markets"]),
    ("India's sovereign wealth fund eyes overseas acquisitions", ["finance", "india", "investment", "economy"]),
    ("Budget 2026 boosts infrastructure allocation 30%", ["economy", "india", "policy", "business"]),
    # World/Policy
    ("G20 nations agree on global minimum corporate tax", ["policy", "global-trade", "economy", "world"]),
    ("UN climate summit sets new 1.4C warming target", ["environment", "policy", "world"]),
    ("US-India tech partnership expands to defense electronics", ["india", "technology", "policy", "world"]),
    ("EU launches €200B green industrial plan", ["environment", "policy", "world", "economy"]),
    ("Middle East sovereign funds pivot to India infrastructure", ["investment", "india", "world", "economy"]),
    # Entertainment/Lifestyle
    ("Bollywood film crosses $200M worldwide opening weekend", ["entertainment", "india", "lifestyle"]),
    ("Indian streaming platforms merge to form major OTT giant", ["entertainment", "technology", "india"]),
    ("Health startup launches AI-powered diagnostics app", ["health", "ai", "startup", "india", "lifestyle"]),
    ("Yoga tourism in India grows 45% in first half of 2026", ["lifestyle", "health", "india", "travel"]),
    ("India food-tech startup raises $50M for cloud kitchens", ["food", "startup", "india", "technology"]),
]

# Pad to 200 items
extra_news = [
    ("Indian fintech startup launches cross-border payments", ["fintech", "startup", "india", "technology"]),
    ("5G rollout reaches 500 Indian cities", ["technology", "india", "telecom"]),
    ("EV adoption in India doubles in Q2 2026", ["technology", "india", "automotive", "environment"]),
    ("India launches second moon mission Chandrayaan-4", ["technology", "india", "space"]),
    ("Cybersecurity investment in India triples post-breach", ["technology", "india", "security"]),
    ("Women entrepreneurship index ranks India 12th globally", ["startup", "india", "business"]),
    ("Renewable energy capacity in India hits 500GW milestone", ["renewable", "solar", "india", "environment"]),
    ("India's pharma exports cross $30B in FY26", ["pharma", "india", "export", "trade"]),
    ("Digital rupee transactions surpass 10M daily", ["finance", "india", "technology"]),
    ("Mumbai emerges as Asia's third-largest startup hub", ["startup", "india", "technology", "business"]),
    ("AI tutoring platform reduces dropout rates in Bihar", ["ai", "education", "india", "technology"]),
    ("India wins bid to host 2030 Commonwealth Games", ["sports", "india", "world"]),
    ("E-sports becomes official medal event in Asian Games", ["sports", "technology", "world"]),
    ("Indian Navy deploys AI-guided patrol vessel", ["technology", "india", "ai", "policy"]),
    ("Global chip shortage eases as India fab comes online", ["semiconductor", "india", "technology", "global-trade"]),
    ("Cotton prices surge 20% on poor monsoon forecast", ["trade", "india", "agriculture"]),
    ("India-Japan bullet train project 60% complete", ["india", "technology", "infrastructure"]),
    ("Online grocery delivery market hits $15B in India", ["food", "technology", "india", "startup"]),
    ("India's space startup ecosystem reaches 150 companies", ["space", "startup", "india", "technology"]),
    ("World Bank upgrades India's ease-of-doing-business rank", ["business", "india", "economy", "policy"]),
    ("Solar energy cheaper than coal for the first time in India", ["solar", "renewable", "india", "energy"]),
    ("AI-based crop monitoring cuts farmer losses 40%", ["ai", "agriculture", "india", "technology"]),
    ("IPL media rights sold for record $6B", ["sports", "cricket", "ipl", "india", "business"]),
    ("India's defence exports hit $3B — a historic record", ["trade", "india", "export", "manufacturing"]),
    ("Microfinance sector sees 25% credit growth in Q1 2026", ["finance", "india", "economy"]),
    ("India announces 100 new smart city projects", ["india", "technology", "policy", "infrastructure"]),
    ("Robotic surgery adoption triples in Indian hospitals", ["health", "technology", "india", "ai"]),
    ("Textile park in Gujarat creates 50,000 jobs", ["textile", "india", "manufacturing", "trade"]),
    ("EdTech platform surpasses 50M learners in India", ["education", "technology", "india", "startup"]),
    ("India's first hydrogen-powered bus fleet launches in Pune", ["technology", "india", "environment", "renewable"]),
    ("Global fashion brands source 30% from India in 2026", ["textile", "trade", "india", "export"]),
    ("India customs modernisation cuts clearance time to 4 hours", ["trade", "india", "logistics", "policy"]),
    ("Crypto regulatory clarity boosts Indian exchange volumes", ["finance", "technology", "india", "policy"]),
    ("India's generative AI market to reach $17B by 2028", ["ai", "india", "technology", "business"]),
    ("Olympic medal hopeful programme supports 500 athletes", ["sports", "india", "olympics"]),
    ("National logistics policy framework finalised", ["logistics", "india", "trade", "policy"]),
    ("Tamil Nadu becomes India's top electronics exporter", ["technology", "india", "export", "semiconductor"]),
    ("Football viewership in India grows 60% year-on-year", ["sports", "football", "india"]),
    ("Telemedicine consultations hit 2M per day in India", ["health", "technology", "india"]),
    ("India joins elite club of nuclear fusion research nations", ["technology", "india", "world", "energy"]),
    ("B2B SaaS startup from Hyderabad goes public on Nasdaq", ["startup", "india", "technology", "business"]),
    ("India's coral reef restoration project wins UN award", ["environment", "india", "world"]),
    ("ISRO satellite internet service rivals Starlink speeds", ["technology", "india", "space"]),
    ("Agri startup raises $30M for precision farming platform", ["agriculture", "startup", "india", "technology"]),
    ("India tightens data localisation laws for AI companies", ["ai", "policy", "india", "technology"]),
    ("Electric two-wheeler sales cross 1M units per month", ["technology", "india", "automotive", "environment"]),
    ("Indian restaurant chain expands to 20 countries", ["food", "india", "lifestyle", "business"]),
    ("Mental health app usage rises 80% among Indian youth", ["health", "india", "lifestyle", "technology"]),
    ("India and UAE sign digital economy partnership", ["technology", "trade", "india", "world"]),
    ("Deep sea mining rights granted to Indian firm", ["india", "world", "environment", "policy"]),
    ("Bangalore overtakes Singapore in startup density", ["startup", "india", "technology", "business"]),
    ("Sports betting legalised in three Indian states", ["sports", "india", "policy", "business"]),
    ("AI-designed drug enters Phase 3 trials in India", ["ai", "health", "india", "technology"]),
    ("India's fastest supercomputer ranks 5th globally", ["technology", "india", "ai", "semiconductor"]),
    ("New airport opens in Tier-2 city Indore", ["india", "infrastructure", "business"]),
    ("Artisan craft exports boosted by government scheme", ["trade", "india", "export", "lifestyle"]),
    ("India wins bronze at World Athletics Championship", ["sports", "athletics", "india"]),
    ("Startup accelerator cohort 2026 includes 40% women founders", ["startup", "india", "business"]),
    ("AI content moderation startup raises $25M", ["ai", "startup", "india", "technology"]),
    ("India overtakes UK in total billionaire count", ["business", "india", "finance", "economy"]),
    ("Healthcare AI reduces misdiagnosis rates by 35%", ["health", "ai", "india", "technology"]),
    ("Green hydrogen project inaugurated in Rajasthan", ["renewable", "energy", "india", "environment"]),
    ("India's fishery exports cross $10B", ["trade", "india", "export", "food"]),
    ("Online education platform raises $100M Series D", ["education", "startup", "india", "technology"]),
    ("India-Vietnam trade ties deepen with new MoU", ["trade", "india", "global-trade", "world"]),
    ("PM inaugurates world's longest rail tunnel", ["india", "infrastructure", "technology"]),
    ("Cricket World Cup 2026 draws 1 billion viewers globally", ["sports", "cricket", "india", "world"]),
    ("Inflation falls to 3.2% — 4-year low in India", ["economy", "india", "finance"]),
    ("Medical device exports from India double in 2 years", ["health", "trade", "india", "export"]),
    ("India's gaming industry to be worth $7B by 2027", ["technology", "entertainment", "india", "startup"]),
    ("Semiconductor design talent pool in India reaches 100,000", ["semiconductor", "india", "technology"]),
    ("India opens borders to digital nomad visa category", ["lifestyle", "india", "policy", "technology"]),
    ("World Bank funds $2B India urban mobility project", ["india", "infrastructure", "economy", "world"]),
    ("AI-powered weather forecasting cuts crop loss in India", ["ai", "agriculture", "india", "technology"]),
    ("India's first commercial thorium reactor approved", ["energy", "india", "technology", "policy"]),
    ("Tea exports from Assam hit record high", ["trade", "india", "export", "food"]),
    ("Biotech startups in Hyderabad triple in 3 years", ["biotech", "startup", "india", "technology"]),
    ("India hosts inaugural Global Semiconductor Summit", ["semiconductor", "india", "technology", "world"]),
    ("SEBI introduces AI-based market surveillance system", ["finance", "india", "ai", "policy"]),
    ("National sports university produces first Olympic medallist", ["sports", "india", "olympics", "education"]),
    ("Rural broadband scheme connects 200,000 villages", ["technology", "india", "policy"]),
    ("Carbon credit market launched on BSE", ["environment", "finance", "india", "policy"]),
    ("India's first quantum internet pilot goes live in Delhi", ["technology", "india", "quantum", "ai"]),
    ("Consumer electronics imports fall as local brands grow", ["technology", "india", "manufacturing", "trade"]),
    ("Indian chef wins Michelin star in London", ["food", "india", "lifestyle", "world"]),
    ("India launches AI safety research institute", ["ai", "policy", "india", "technology"]),
    ("Organic produce exports from India grow 55%", ["food", "trade", "india", "export", "environment"]),
    ("Stock market regulator approves tokenised securities", ["finance", "india", "technology", "policy"]),
    ("India-UK digital trade corridor operational from July", ["trade", "india", "technology", "world"]),
    ("Maharashtra leads states in renewable energy output", ["renewable", "india", "energy", "environment"]),
    ("AI chatbot resolves 70% of Indian tax queries", ["ai", "india", "technology", "finance"]),
    ("Kabaddi World Cup: India wins sixth consecutive title", ["sports", "kabaddi", "india"]),
    ("Indian space tourism flight scheduled for 2027", ["space", "india", "technology", "lifestyle"]),
    ("B2B marketplace for industrial goods raises $80M", ["b2b", "startup", "india", "trade"]),
    ("India's telematics insurance market triples", ["technology", "india", "finance"]),
    ("National AI curriculum rolled out in schools", ["ai", "education", "india", "policy"]),
    ("Film industry adopts AI for VFX; cost drops 60%", ["ai", "entertainment", "india", "technology"]),
    ("Tier-2 startup cities get ₹500Cr government boost", ["startup", "india", "policy", "business"]),
    ("Yoga app surpasses 30M users globally", ["lifestyle", "health", "india", "technology"]),
    ("India signs defence electronics deal with Japan", ["technology", "india", "world", "policy"]),
    ("Textile recycling startup raises $15M in seed round", ["textile", "environment", "startup", "india"]),
    ("India's unemployment rate drops to 4.2%", ["economy", "india", "business"]),
    ("Open source AI framework from IISc gains global traction", ["ai", "technology", "india", "startup"]),
    ("Digital health record system now live in 28 states", ["health", "india", "technology", "policy"]),
    ("India automotive exports cross $25B milestone", ["automotive", "india", "export", "trade"]),
    ("Smart meter rollout: 100M households connected", ["technology", "india", "energy", "infrastructure"]),
    ("Anti-drone tech startup from DRDO spin-off raises $40M", ["technology", "startup", "india", "ai"]),
    ("India's coffee exports find premium buyers in Scandinavia", ["food", "trade", "india", "export"]),
    ("India ranked top 10 globally for quality of life improvements", ["india", "lifestyle", "economy", "world"]),
    ("Electric aircraft startup targets regional routes in India", ["technology", "india", "startup", "environment"]),
    ("Agri-fintech bridges credit gap for 5M smallholders", ["agriculture", "finance", "india", "startup"]),
    ("India wins data localisation dispute at WTO", ["policy", "india", "technology", "world", "trade"]),
    ("Drone delivery reaches 50,000 pincodes in India", ["technology", "india", "startup", "logistics"]),
    ("IPO pipeline for FY27 includes 60 tech companies", ["startup", "india", "finance", "technology"]),
    ("Global luxury brands expand India presence significantly", ["lifestyle", "india", "business"]),
    ("Wearable health device market in India grows 90%", ["health", "technology", "india", "lifestyle"]),
    ("India MSME export council signs deal with Amazon Global", ["trade", "india", "export", "b2b"]),
    ("Kerala model of community farming inspires 10 nations", ["agriculture", "india", "world", "environment"]),
    ("AI-based legal assistant passes bar exam in India", ["ai", "india", "technology"]),
    ("Indian sprinter runs sub-10s 100m for first time", ["sports", "athletics", "india", "olympics"]),
    ("Nirmala Sitharaman presents third consecutive budget surplus", ["finance", "india", "economy", "policy"]),
    ("India's cold chain infrastructure investment hits $3B", ["logistics", "india", "trade", "food"]),
    ("AI-powered solar grid management saves ₹2,000Cr", ["ai", "solar", "india", "energy", "technology"]),
    ("Indian video game wins BAFTA award", ["entertainment", "india", "technology", "world"]),
    ("Padel tennis courts boom across Indian metros", ["sports", "lifestyle", "india"]),
    ("India's largest electric bus fleet deployed in Delhi", ["technology", "india", "environment", "infrastructure"]),
    ("Precision agriculture startup partners with 1M farmers", ["agriculture", "startup", "india", "technology"]),
    ("India launches carbon border adjustment mechanism", ["environment", "india", "trade", "policy"]),
    ("Cricket streaming rights spark bidding war among OTTs", ["sports", "cricket", "india", "technology"]),
    ("Indian diplomat heads UN climate change committee", ["world", "india", "environment", "policy"]),
    ("FMCG giant launches AI-personalised product lines", ["ai", "technology", "india", "business"]),
    ("Microchip design school opens in Mysore", ["semiconductor", "india", "education", "technology"]),
    ("India's carbon neutrality target moved to 2045", ["environment", "india", "policy", "world"]),
    ("Social commerce startup hits $1B GMV in 12 months", ["startup", "india", "technology", "business"]),
    ("PM unveils 10-year digital infrastructure master plan", ["technology", "india", "policy"]),
    ("Indian hip-hop artist wins Grammy for first time", ["entertainment", "india", "lifestyle", "world"]),
    ("Global food brands reformulate for Indian palate", ["food", "india", "business", "lifestyle"]),
    ("India surpasses Germany as world's 3rd largest economy", ["economy", "india", "world", "business"]),
    ("Trade facilitation index: India climbs to 8th globally", ["trade", "india", "world", "policy"]),
    ("New drug from Indian lab targets drug-resistant TB", ["health", "india", "technology", "world"]),
    ("AI startup builds real-time sign language interpreter", ["ai", "startup", "india", "technology"]),
    ("Heritage tourism surges: 50M visits to monuments in H1", ["lifestyle", "india", "travel", "world"]),
    ("India's IT services export outlook raised to $320B", ["technology", "india", "export", "business"]),
]

news_items = []
all_news = news_templates + extra_news
for idx, (title, labels) in enumerate(all_news[:200], start=1):
    news_items.append({
        "ItemId": f"news_{idx:03d}",
        "Labels": labels,
        "Categories": ["news"],
        "Timestamp": rand_ts(1, 23),
        "Comment": title,
        "IsHidden": False
    })

with open(os.path.join(BASE_DIR, "news_items.json"), "w") as f:
    json.dump(news_items, f, indent=2)
print(f"Generated {len(news_items)} news items")

# ── PRODUCT ITEMS ────────────────────────────────────────────────────────────

products_data = [
    # Electronics
    ("Samsung Galaxy S26 Ultra 512GB", ["electronics", "smartphone", "premium", "samsung"]),
    ("Apple iPhone 16 Pro Max 256GB", ["electronics", "smartphone", "premium", "apple"]),
    ("Sony WH-1000XM6 Noise Cancelling Headphones", ["electronics", "audio", "headphones", "sony"]),
    ("Dell XPS 15 OLED Laptop", ["electronics", "laptop", "premium", "dell"]),
    ("LG 55\" 4K OLED Smart TV", ["electronics", "tv", "home-entertainment", "lg"]),
    ("Bose SoundLink Max Speaker", ["electronics", "audio", "speaker", "bose"]),
    ("Apple MacBook Pro M4 14\"", ["electronics", "laptop", "premium", "apple"]),
    ("OnePlus 13 Pro 5G", ["electronics", "smartphone", "oneplus"]),
    ("Asus ROG Strix Gaming Laptop", ["electronics", "laptop", "gaming", "asus"]),
    ("Amazon Echo Hub Smart Display", ["electronics", "smart-home", "amazon"]),
    ("Xiaomi Smart TV Pro 43\"", ["electronics", "tv", "xiaomi", "budget"]),
    ("Realme Buds Air 5 TWS", ["electronics", "audio", "earbuds", "budget"]),
    ("Nikon Z8 Mirrorless Camera", ["electronics", "camera", "photography", "nikon"]),
    ("GoPro Hero 13 Black", ["electronics", "camera", "action", "gopro"]),
    ("iPad Pro 13\" M4 WiFi", ["electronics", "tablet", "premium", "apple"]),
    ("Microsoft Surface Pro 10", ["electronics", "tablet", "laptop", "microsoft"]),
    ("Logitech MX Master 3S Mouse", ["electronics", "accessories", "logitech"]),
    ("Samsung 27\" QHD Gaming Monitor", ["electronics", "monitor", "gaming", "samsung"]),
    ("TP-Link Deco WiFi 7 Mesh System", ["electronics", "networking", "smart-home"]),
    ("Dyson V15 Detect Vacuum", ["electronics", "home-appliances", "dyson"]),
    # Fashion
    ("Levi's 501 Original Fit Jeans", ["fashion", "denim", "levis", "casual"]),
    ("Nike Air Max 2026 Running Shoes", ["fashion", "footwear", "sports", "nike"]),
    ("Adidas Ultraboost 25", ["fashion", "footwear", "sports", "adidas"]),
    ("Zara Structured Blazer", ["fashion", "formal", "zara", "office"]),
    ("H&M Relaxed Linen Shirt", ["fashion", "casual", "hm", "summer"]),
    ("Mango Pleated Midi Skirt", ["fashion", "women", "mango"]),
    ("Puma RS-X Trail Sneakers", ["fashion", "footwear", "casual", "puma"]),
    ("Allen Solly Slim Fit Chinos", ["fashion", "formal", "casual", "men"]),
    ("Fabindia Khadi Cotton Kurta", ["fashion", "ethnic", "india", "fabindia"]),
    ("Woodland Trek Boots", ["fashion", "footwear", "outdoor", "woodland"]),
    ("Sabyasachi Printed Saree", ["fashion", "ethnic", "premium", "india", "women"]),
    ("W for Woman Printed Kurti Set", ["fashion", "ethnic", "women", "india"]),
    ("Arrow Formal Oxford Shirt", ["fashion", "formal", "men", "office"]),
    ("Reebok Classic Leather Shoes", ["fashion", "footwear", "casual", "reebok"]),
    ("Biba Embroidered Anarkali", ["fashion", "ethnic", "women", "india", "biba"]),
    # Home Appliances
    ("LG 8kg Inverter Washing Machine", ["home-appliances", "lg", "laundry"]),
    ("Bosch Series 6 Dishwasher", ["home-appliances", "bosch", "kitchen"]),
    ("Haier 340L Double Door Refrigerator", ["home-appliances", "haier", "kitchen"]),
    ("Whirlpool 1.5T 5 Star AC", ["home-appliances", "whirlpool", "cooling"]),
    ("Philips Air Fryer XXL", ["home-appliances", "philips", "kitchen", "cooking"]),
    ("Instant Pot Duo 7-in-1", ["home-appliances", "cooking", "kitchen"]),
    ("Bajaj Majesty OTG 35L", ["home-appliances", "cooking", "kitchen", "bajaj"]),
    ("Prestige Induction Cooktop", ["home-appliances", "cooking", "kitchen", "prestige"]),
    ("Dyson Purifier Cool Auto React", ["home-appliances", "air-purifier", "dyson"]),
    ("Eureka Forbes AquaSure RO+UV", ["home-appliances", "water-purifier", "kitchen"]),
    # Books
    ("The Coming Wave – Mustafa Suleyman", ["books", "technology", "ai", "nonfiction"]),
    ("Chip War – Chris Miller", ["books", "technology", "semiconductor", "nonfiction"]),
    ("India 2047 – Vision for a $30T Economy", ["books", "india", "economy", "nonfiction"]),
    ("Atomic Habits – James Clear", ["books", "self-help", "productivity", "nonfiction"]),
    ("The Lean Startup – Eric Ries", ["books", "startup", "business", "nonfiction"]),
    ("Zero to One – Peter Thiel", ["books", "startup", "business", "nonfiction"]),
    ("Sapiens – Yuval Noah Harari", ["books", "history", "nonfiction", "bestseller"]),
    ("The Psychology of Money – Morgan Housel", ["books", "finance", "nonfiction", "bestseller"]),
    ("Deep Work – Cal Newport", ["books", "productivity", "self-help", "nonfiction"]),
    ("No Rules Rules – Reed Hastings", ["books", "business", "startup", "nonfiction"]),
    # Accessories
    ("Fossil Gen 6 Hybrid Smartwatch", ["accessories", "watch", "fossil", "wearable"]),
    ("boAt Wave Neo 2 Smartwatch", ["accessories", "watch", "boat", "budget", "wearable"]),
    ("Wildcraft 30L Hiking Backpack", ["accessories", "outdoor", "backpack", "wildcraft"]),
    ("American Tourister Spinner Trolley", ["accessories", "travel", "luggage"]),
    ("Fossil Bifold Leather Wallet", ["accessories", "wallet", "leather", "fossil"]),
    ("Ray-Ban Wayfarer Sunglasses", ["accessories", "sunglasses", "rayban", "premium"]),
    ("Noise ColorFit Ultra 3 Watch", ["accessories", "watch", "wearable", "budget"]),
    ("VIP Overnighter Travel Bag", ["accessories", "travel", "luggage", "vip"]),
    ("Puma Sport Backpack 30L", ["accessories", "backpack", "sports", "puma"]),
    ("Caprese Sling Bag", ["accessories", "bag", "women", "caprese"]),
    # Extra to reach 100
    ("Panasonic 55\" 4K HDR TV", ["electronics", "tv", "panasonic"]),
    ("Nothing Phone (3a) Pro", ["electronics", "smartphone", "nothing"]),
    ("JBL Flip 7 Portable Speaker", ["electronics", "audio", "speaker", "jbl"]),
    ("Honor Magic7 Pro 5G", ["electronics", "smartphone", "honor"]),
    ("Vivo X200 Pro", ["electronics", "smartphone", "vivo"]),
    ("Oppo Find X8 Ultra", ["electronics", "smartphone", "oppo"]),
    ("Amazfit GTR 5 Smartwatch", ["accessories", "watch", "wearable", "amazfit"]),
    ("Titan Raga Gold Watch", ["accessories", "watch", "titan", "premium", "women"]),
    ("Casio G-Shock GA-2100", ["accessories", "watch", "casio", "men"]),
    ("Canon EOS R8 Mirrorless", ["electronics", "camera", "photography", "canon"]),
    ("FabAlley Floral Wrap Dress", ["fashion", "women", "summer", "casual"]),
    ("USPA Polo T-Shirt Pack of 3", ["fashion", "casual", "men"]),
    ("Crocs Classic Clog", ["fashion", "footwear", "casual", "crocs"]),
    ("New Balance 530 Retro Shoes", ["fashion", "footwear", "casual", "new-balance"]),
    ("Camper Nepal Derby Shoes", ["fashion", "footwear", "formal", "camper"]),
    ("IFB 6.5kg Fully Automatic Washer", ["home-appliances", "laundry", "ifb"]),
    ("Pigeon by Stovekraft Pressure Cooker", ["home-appliances", "kitchen", "cooking"]),
    ("Kent Grand Plus Water Purifier", ["home-appliances", "water-purifier", "kitchen"]),
    ("Crompton 48W Ceiling Fan", ["home-appliances", "cooling", "crompton"]),
    ("Havells Stabilizer 4kVA", ["home-appliances", "stabilizer", "havells"]),
    ("The Great Indian Kitchen – Nisha Madhulika", ["books", "food", "india", "nonfiction"]),
    ("The God of Small Things – Arundhati Roy", ["books", "fiction", "india", "literature"]),
    ("India After Gandhi – Ramachandra Guha", ["books", "india", "history", "nonfiction"]),
    ("The Innovators – Walter Isaacson", ["books", "technology", "history", "nonfiction"]),
    ("Poor Charlie's Almanack", ["books", "finance", "business", "nonfiction"]),
    ("Wildcraft Rucksack 60L", ["accessories", "outdoor", "backpack", "wildcraft"]),
    ("Beetle Classic Belt Combo Pack", ["accessories", "belt", "men", "casual"]),
    ("F Gear Gym Bag 37L", ["accessories", "sports", "bag", "fitness"]),
    ("Samsonite Prodigy Backpack", ["accessories", "backpack", "travel", "samsonite"]),
    ("Da Milano Leather Handbag", ["accessories", "bag", "women", "premium", "da-milano"]),
]

product_items = []
for idx, (title, labels) in enumerate(products_data[:100], start=1):
    product_items.append({
        "ItemId": f"prod_{idx:03d}",
        "Labels": labels,
        "Categories": ["products"],
        "Comment": title,
        "IsHidden": False
    })

with open(os.path.join(BASE_DIR, "product_items.json"), "w") as f:
    json.dump(product_items, f, indent=2)
print(f"Generated {len(product_items)} product items")

# ── EXPO BOOTHS ──────────────────────────────────────────────────────────────

expo_data = [
    ("SunRise Solar – Photovoltaic Modules Manufacturer", ["solar", "renewable", "energy", "export", "manufacturing"]),
    ("GreenWatt Energy Solutions – Off-grid Solar Systems", ["solar", "renewable", "energy", "b2b"]),
    ("Tata Power Solar – Large Scale Solar Projects", ["solar", "renewable", "energy", "india", "tata"]),
    ("WindForce India – Wind Turbine Components", ["wind", "renewable", "energy", "manufacturing"]),
    ("HydroGen Pro – Green Hydrogen Production", ["hydrogen", "renewable", "energy", "b2b"]),
    ("EnergoLink – Smart Grid & Energy Storage", ["energy", "technology", "smart-grid", "storage"]),
    ("PV Module Hub – Solar Panel Trading Platform", ["solar", "trade", "b2b", "export"]),
    ("BioFuel India – Agricultural Waste to Biofuel", ["biofuel", "renewable", "agriculture", "environment"]),
    ("SolarBank Finance – Solar Project Financing", ["solar", "finance", "renewable", "b2b"]),
    ("RenewTech Labs – R&D Solar Efficiency", ["solar", "technology", "research", "renewable"]),
    ("WeaveIndia Textiles – Premium Cotton Fabrics", ["textile", "cotton", "export", "india", "b2b"]),
    ("SilkRoute Exports – Handloom & Silk", ["textile", "silk", "export", "india", "handloom"]),
    ("Texport Industries – Technical Textiles", ["textile", "technical", "manufacturing", "export"]),
    ("JewelThread – Zari & Embroidery Works", ["textile", "embroidery", "export", "india"]),
    ("KnitWell Hosiery – Knitwear Manufacturer", ["textile", "knitwear", "manufacturing", "export"]),
    ("Arvind Mills – Denim & Apparel Fabrics", ["textile", "denim", "manufacturing", "india"]),
    ("FiberFlex Nonwovens – Nonwoven Fabric Exporter", ["textile", "nonwoven", "export", "b2b"]),
    ("EcoYarn India – Recycled Yarn Manufacturer", ["textile", "recycled", "environment", "export"]),
    ("BoldPrint Garments – Bulk Garment Supplier", ["textile", "garment", "export", "b2b"]),
    ("CottonKing Exports – Organic Cotton Trader", ["textile", "cotton", "organic", "export", "trade"]),
    ("Chemix India – Industrial Chemicals Exporter", ["chemicals", "export", "b2b", "manufacturing"]),
    ("PharmaChem Solutions – API & Intermediates", ["chemicals", "pharma", "export", "b2b"]),
    ("AgroStar Chemicals – Crop Protection Chemicals", ["chemicals", "agriculture", "export", "india"]),
    ("CleanChem Coatings – Specialty Coatings", ["chemicals", "specialty", "manufacturing", "export"]),
    ("DyeStar India – Textile Dyes & Auxiliaries", ["chemicals", "textile", "export", "manufacturing"]),
    ("PetroBase Lubricants – Industrial Lubricants", ["chemicals", "industrial", "manufacturing", "export"]),
    ("NanoTech Pigments – Advanced Pigment Exporter", ["chemicals", "technology", "export", "b2b"]),
    ("BioSynth India – Biochemicals Producer", ["chemicals", "biotech", "export", "b2b"]),
    ("FoodChem Additives – Food Grade Ingredients", ["chemicals", "food", "export", "b2b"]),
    ("PolymersPlus – Engineering Plastics & Polymers", ["chemicals", "plastics", "manufacturing", "export"]),
    ("SpiceTrade India – Premium Spice Exporter", ["food", "spices", "export", "india", "trade"]),
    ("GrainLink – Basmati Rice & Cereals Export", ["food", "rice", "export", "india", "trade"]),
    ("FreshPack – Processed Fruits & Vegetables", ["food", "processed", "export", "india"]),
    ("AquaHarvest – Marine Products Exporter", ["food", "seafood", "export", "india"]),
    ("DairyGlobe – Dairy Products Manufacturer", ["food", "dairy", "manufacturing", "export"]),
    ("TeaFirst – Assam & Darjeeling Tea Exporter", ["food", "tea", "export", "india", "b2b"]),
    ("OilPress India – Cold-Pressed Oil Exporter", ["food", "oil", "export", "india", "organic"]),
    ("OrganicRoots – Certified Organic Food Exporter", ["food", "organic", "export", "india", "b2b"]),
    ("NutriPack – Nutraceuticals & Health Foods", ["food", "health", "export", "b2b"]),
    ("BeverageBridge – Ready-to-Drink Beverages", ["food", "beverages", "export", "b2b"]),
    ("TechAxis – Precision Engineering Parts", ["technology", "manufacturing", "export", "b2b"]),
    ("SoftGlobal – Custom Software Solutions", ["technology", "software", "b2b", "export"]),
    ("AI Foundry India – AI Services for Enterprises", ["technology", "ai", "b2b", "india"]),
    ("CyberShield India – Cybersecurity Services", ["technology", "security", "b2b", "india"]),
    ("IoTBridge – Industrial IoT Solutions", ["technology", "iot", "manufacturing", "b2b"]),
    ("AutoBot Systems – Factory Automation", ["technology", "robotics", "manufacturing", "b2b"]),
    ("EduTech Global – Online Learning Platforms", ["technology", "education", "b2b", "export"]),
    ("HealthTech India – Digital Health Solutions", ["technology", "health", "b2b", "india"]),
    ("CloudNine Services – Managed Cloud Solutions", ["technology", "cloud", "b2b", "export"]),
    ("BlockChainX – Supply Chain Traceability", ["technology", "blockchain", "supply-chain", "b2b"]),
]

expo_booths = []
for idx, (title, labels) in enumerate(expo_data[:50], start=1):
    expo_booths.append({
        "ItemId": f"booth_{idx:03d}",
        "Labels": labels,
        "Categories": ["expo_booth"],
        "Comment": title,
        "IsHidden": False
    })

with open(os.path.join(BASE_DIR, "expo_booths.json"), "w") as f:
    json.dump(expo_booths, f, indent=2)
print(f"Generated {len(expo_booths)} expo booths")

# ── FEEDBACK ─────────────────────────────────────────────────────────────────

tech_user_ids  = [f"u_{i:03d}" for i in range(1, 31)]  + ["u_ai_reader"]
trade_user_ids = [f"u_{i:03d}" for i in range(31, 51)] + ["u_trade_reader"]
sports_user_ids= [f"u_{i:03d}" for i in range(51, 66)] + ["u_sports_reader"]
biz_user_ids   = [f"u_{i:03d}" for i in range(66, 81)]
life_user_ids  = [f"u_{i:03d}" for i in range(81, 91)]
buyer_ids      = [f"u_{i:03d}" for i in range(91, 101)] + [
    "buyer_energy_001", "buyer_textile_001", "buyer_energy_009", "buyer_import_014"
]

# Feedback types with weights
fb_types = ["view","view","view","view","view","click","click","read","read","read","like","bookmark","share","purchase","inquiry","meeting_request"]

news_ids    = [f"news_{i:03d}" for i in range(1, 201)]
prod_ids    = [f"prod_{i:03d}" for i in range(1, 101)]
booth_ids   = [f"booth_{i:03d}" for i in range(1, 51)]

# Tech/AI news: news_001-news_015 (first 15 = AI/tech)
tech_news   = [f"news_{i:03d}" for i in range(1, 16)]
trade_news  = [f"news_{i:03d}" for i in range(16, 26)]
sports_news = [f"news_{i:03d}" for i in range(26, 36)]
biz_news    = [f"news_{i:03d}" for i in range(36, 51)]

energy_booths  = [f"booth_{i:03d}" for i in range(1, 11)]
textile_booths = [f"booth_{i:03d}" for i in range(11, 21)]
chem_booths    = [f"booth_{i:03d}" for i in range(21, 31)]
food_booths    = [f"booth_{i:03d}" for i in range(31, 41)]
tech_booths    = [f"booth_{i:03d}" for i in range(41, 51)]

electronics = [f"prod_{i:03d}" for i in range(1, 21)]
fashion_p   = [f"prod_{i:03d}" for i in range(21, 36)]
home_appl   = [f"prod_{i:03d}" for i in range(36, 51)]
books_p     = [f"prod_{i:03d}" for i in range(51, 61)]
access_p    = [f"prod_{i:03d}" for i in range(61, 76)]

feedback = []

def add_fb(user_id, item_ids, count, types=None):
    if types is None:
        types = fb_types
    for _ in range(count):
        feedback.append({
            "FeedbackType": pick(types),
            "UserId": user_id,
            "ItemId": pick(item_ids),
            "Timestamp": rand_ts()
        })

# Tech users → tech news + electronics + some books
for uid in tech_user_ids:
    add_fb(uid, tech_news, random.randint(10, 20))
    add_fb(uid, news_ids,  random.randint(3, 6))
    add_fb(uid, electronics, random.randint(3, 7))
    add_fb(uid, books_p,     random.randint(1, 3))

# Trade users → trade news + expo booths (trade)
for uid in trade_user_ids:
    add_fb(uid, trade_news,   random.randint(8, 15))
    add_fb(uid, news_ids,     random.randint(2, 5))
    add_fb(uid, textile_booths, random.randint(3, 6))
    add_fb(uid, chem_booths,    random.randint(2, 4))
    add_fb(uid, food_booths,    random.randint(1, 3))

# Sports users → sports news + fashion/accessories
for uid in sports_user_ids:
    add_fb(uid, sports_news, random.randint(8, 18))
    add_fb(uid, news_ids,    random.randint(2, 5))
    add_fb(uid, fashion_p,   random.randint(3, 6))
    add_fb(uid, access_p,    random.randint(2, 4))

# Business users → biz news + books + finance products
for uid in biz_user_ids:
    add_fb(uid, biz_news, random.randint(8, 14))
    add_fb(uid, news_ids, random.randint(2, 5))
    add_fb(uid, books_p,  random.randint(3, 6))
    add_fb(uid, electronics, random.randint(2, 4))

# Lifestyle users → all news + all products
for uid in life_user_ids:
    add_fb(uid, news_ids,   random.randint(4, 8))
    add_fb(uid, fashion_p,  random.randint(3, 7))
    add_fb(uid, home_appl,  random.randint(3, 6))
    add_fb(uid, access_p,   random.randint(2, 4))

# Buyers → expo booths
add_fb("buyer_energy_001", energy_booths, 12, ["view","click","inquiry","meeting_request"])
add_fb("buyer_energy_009", energy_booths, 10, ["view","click","inquiry"])
add_fb("buyer_textile_001", textile_booths, 10, ["view","click","inquiry","meeting_request"])
add_fb("buyer_import_014", chem_booths, 8, ["view","click","inquiry"])

for uid in buyer_ids[:10]:
    add_fb(uid, booth_ids,     random.randint(5, 10))
    add_fb(uid, trade_news,    random.randint(2, 4))
    add_fb(uid, energy_booths, random.randint(2, 4))

# Mixed users cross-category
all_users = tech_user_ids + trade_user_ids + sports_user_ids + biz_user_ids
for uid in random.sample(all_users, 30):
    add_fb(uid, news_ids,  random.randint(2, 4))
    add_fb(uid, prod_ids,  random.randint(1, 3))

print(f"Generated {len(feedback)} feedback events")

# Shuffle
random.shuffle(feedback)

with open(os.path.join(BASE_DIR, "feedback.json"), "w") as f:
    json.dump(feedback, f, indent=2)
print(f"Saved {len(feedback)} feedback events to feedback.json")
print("All data files generated successfully!")
