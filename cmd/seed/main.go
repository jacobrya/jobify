package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/abzalserikbay/jobify/pkg/hasher"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("ping: %v", err)
	}

	h := hasher.New()
	pwHash, err := h.Hash("Test1234!")
	if err != nil {
		log.Fatalf("hash: %v", err)
	}

	rng := rand.New(rand.NewSource(42))

	fmt.Println("seeding users...")
	userIDs := seedUsers(ctx, db, pwHash)

	fmt.Println("seeding jobs...")
	jobIDs := seedJobs(ctx, db)

	fmt.Println("seeding applications...")
	seedApplications(ctx, db, userIDs, jobIDs, rng)

	fmt.Println("seeding saved jobs...")
	seedSavedJobs(ctx, db, userIDs, jobIDs, rng)

	fmt.Printf("done. users=%d jobs=%d\n", len(userIDs), len(jobIDs))
}

// ─── users ────────────────────────────────────────────────────────────────────

type userData struct {
	email string
	role  string
	name  string
	bio   string
	skills         []string
	experienceYears int
	salaryMin      int
	salaryMax      int
	remoteOnly     bool
	githubURL      string
}

var users = []userData{
	{
		email: "admin@jobify.dev", role: "admin",
		name: "Admin", bio: "Platform administrator.",
		skills: []string{"go", "postgres", "redis", "docker"},
		experienceYears: 10, salaryMin: 0, salaryMax: 0,
		remoteOnly: false, githubURL: "",
	},
	{
		email: "alex.kim@gmail.com", role: "developer",
		name: "Alex Kim",
		bio:  "Backend engineer obsessed with Go and distributed systems. Previously at Cloudflare.",
		skills:          []string{"go", "postgres", "redis", "kafka", "grpc", "docker", "kubernetes"},
		experienceYears: 5, salaryMin: 5000, salaryMax: 9000,
		remoteOnly: true, githubURL: "https://github.com/alexkim-dev",
	},
	{
		email: "maria.chen@outlook.com", role: "developer",
		name: "Maria Chen",
		bio:  "ML engineer with a background in NLP and computer vision. Love Python and PyTorch.",
		skills:          []string{"python", "pytorch", "tensorflow", "mlflow", "sql", "docker", "kubernetes"},
		experienceYears: 4, salaryMin: 6000, salaryMax: 11000,
		remoteOnly: true, githubURL: "https://github.com/mariachen-ml",
	},
	{
		email: "david.park@gmail.com", role: "developer",
		name: "David Park",
		bio:  "Frontend lead. 4 years building design systems and complex SPAs with React and TypeScript.",
		skills:          []string{"react", "typescript", "nextjs", "graphql", "tailwind", "figma"},
		experienceYears: 4, salaryMin: 4000, salaryMax: 7500,
		remoteOnly: false, githubURL: "https://github.com/davidpark-ui",
	},
	{
		email: "sarah.jones@gmail.com", role: "developer",
		name: "Sarah Jones",
		bio:  "SRE / DevOps engineer. I make systems reliable at scale. Terraform, k8s, observability.",
		skills:          []string{"kubernetes", "terraform", "prometheus", "grafana", "aws", "go", "python", "linux"},
		experienceYears: 7, salaryMin: 7000, salaryMax: 13000,
		remoteOnly: true, githubURL: "https://github.com/sre-sarah",
	},
	{
		email: "nikita.volkov@gmail.com", role: "developer",
		name: "Nikita Volkov",
		bio:  "Systems programmer. Rust for performance-critical code, Go for services. Ex-Yandex.",
		skills:          []string{"rust", "go", "c++", "wasm", "linux", "postgres"},
		experienceYears: 3, salaryMin: 5000, salaryMax: 9500,
		remoteOnly: true, githubURL: "https://github.com/nvolkov",
	},
	{
		email: "amir.hassan@gmail.com", role: "developer",
		name: "Amir Hassan",
		bio:  "Full-stack senior. Node.js backend, React frontend. 7 years shipping products.",
		skills:          []string{"nodejs", "typescript", "react", "postgres", "mongodb", "aws", "docker"},
		experienceYears: 7, salaryMin: 5500, salaryMax: 9000,
		remoteOnly: false, githubURL: "https://github.com/amirhassan",
	},
	{
		email: "priya.sharma@gmail.com", role: "developer",
		name: "Priya Sharma",
		bio:  "Data engineer building pipelines that actually work. dbt, Spark, Airflow, Snowflake.",
		skills:          []string{"python", "sql", "dbt", "spark", "airflow", "snowflake", "kafka"},
		experienceYears: 5, salaryMin: 5500, salaryMax: 10000,
		remoteOnly: true, githubURL: "https://github.com/priya-data",
	},
	{
		email: "lucas.martin@gmail.com", role: "developer",
		name: "Lucas Martin",
		bio:  "Java / Spring Boot veteran. 8 years building enterprise microservices. Now exploring Go.",
		skills:          []string{"java", "spring", "postgres", "kafka", "docker", "kubernetes", "go"},
		experienceYears: 8, salaryMin: 6000, salaryMax: 10000,
		remoteOnly: false, githubURL: "https://github.com/lucas-martin",
	},
	{
		email: "yuki.tanaka@gmail.com", role: "developer",
		name: "Yuki Tanaka",
		bio:  "iOS engineer. Swift + SwiftUI. Shipped 5 apps on the App Store. Ex-LINE.",
		skills:          []string{"swift", "swiftui", "xcode", "ios", "objc", "firebase"},
		experienceYears: 4, salaryMin: 4500, salaryMax: 8000,
		remoteOnly: false, githubURL: "https://github.com/yukitanaka-ios",
	},
	{
		email: "emma.wilson@gmail.com", role: "developer",
		name: "Emma Wilson",
		bio:  "Android developer. Kotlin, Compose, clean architecture. 5 years in mobile.",
		skills:          []string{"kotlin", "android", "jetpack compose", "coroutines", "firebase", "java"},
		experienceYears: 5, salaryMin: 5000, salaryMax: 8500,
		remoteOnly: true, githubURL: "https://github.com/emmawilson-android",
	},
	{
		email: "carlos.garcia@gmail.com", role: "developer",
		name: "Carlos Garcia",
		bio:  "Python backend dev. Django, FastAPI, async Python. 6 years in fintech startups.",
		skills:          []string{"python", "django", "fastapi", "postgres", "redis", "celery", "docker"},
		experienceYears: 6, salaryMin: 5000, salaryMax: 9000,
		remoteOnly: false, githubURL: "https://github.com/cgarcia-py",
	},
}

func seedUsers(ctx context.Context, db *pgxpool.Pool, pwHash string) []uuid.UUID {
	var ids []uuid.UUID
	for _, u := range users {
		id := uuid.New()

		var existing uuid.UUID
		_ = db.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, u.email).Scan(&existing)
		if existing != uuid.Nil {
			ids = append(ids, existing)
			continue
		}

		_, err := db.Exec(ctx,
			`INSERT INTO users (id, email, password, role) VALUES ($1,$2,$3,$4)`,
			id, u.email, pwHash, u.role,
		)
		if err != nil {
			log.Printf("user %s: %v", u.email, err)
			continue
		}

		if u.role == "developer" {
			_, err = db.Exec(ctx,
				`INSERT INTO developer_profiles (id, user_id, name, bio, skills, experience_years, salary_min, salary_max, remote_only, github_url)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
				 ON CONFLICT (user_id) DO NOTHING`,
				uuid.New(), id, u.name, u.bio, u.skills,
				u.experienceYears, u.salaryMin, u.salaryMax, u.remoteOnly, u.githubURL,
			)
			if err != nil {
				log.Printf("profile %s: %v", u.email, err)
			}
		}

		ids = append(ids, id)
		fmt.Printf("  + %s (%s)\n", u.email, u.role)
	}
	return ids
}

// ─── jobs ─────────────────────────────────────────────────────────────────────

type jobData struct {
	sourceID    string
	title       string
	company     string
	description string
	skills      []string
	salaryMin   int
	salaryMax   int
	isRemote    bool
	location    string
	jobType     string
	url         string
}

var jobs = []jobData{
	// ── Go / Backend ──────────────────────────────────────────────────────────
	{
		sourceID: "seed_001", title: "Senior Go Engineer", company: "Cloudflare",
		description: "Join Cloudflare's core network team to build the systems that protect and accelerate millions of websites. You'll work on our globally distributed edge network, improving performance and reliability at massive scale.\n\nResponsibilities:\n- Design and implement high-throughput Go services handling millions of req/s\n- Collaborate with the network engineering team on packet processing pipelines\n- Contribute to open-source projects like Pingora\n- Participate in on-call rotation and incident response\n\nRequirements:\n- 4+ years of Go experience\n- Strong understanding of networking (TCP/IP, TLS, HTTP/2, QUIC)\n- Experience with distributed systems and high-availability architectures\n- Familiarity with Linux internals and eBPF is a plus",
		skills:      []string{"go", "distributed systems", "networking", "linux", "kubernetes"},
		salaryMin: 10000, salaryMax: 18000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://cloudflare.com/careers",
	},
	{
		sourceID: "seed_002", title: "Backend Engineer – Payments", company: "Stripe",
		description: "Stripe's Payments team is responsible for the core primitives that power billions of dollars in transactions. We're looking for a backend engineer to help design and build the next generation of our payment processing infrastructure.\n\nYou'll work on:\n- Core payment flows: charges, refunds, disputes\n- Idempotency and consistency at scale\n- API design and versioning strategy\n- Fraud prevention integrations\n\nStack: Ruby, Go, Java, PostgreSQL, Kafka, Hadoop",
		skills:      []string{"go", "ruby", "postgres", "kafka", "distributed systems"},
		salaryMin: 12000, salaryMax: 22000, isRemote: true, location: "USA, Canada",
		jobType: "full_time", url: "https://stripe.com/jobs",
	},
	{
		sourceID: "seed_003", title: "Go Platform Engineer", company: "HashiCorp",
		description: "HashiCorp builds the tools that companies use to provision, secure, connect, and run infrastructure. We're hiring a platform engineer to work on Terraform Cloud's backend services.\n\nKey responsibilities:\n- Build and maintain internal platform APIs in Go\n- Design event-driven systems using Kafka and NATS\n- Improve developer experience through better tooling and automation\n- Contribute to Terraform provider SDKs\n\nWe are fully remote and have a strong async-first culture.",
		skills:      []string{"go", "terraform", "kubernetes", "aws", "grpc", "kafka"},
		salaryMin: 9000, salaryMax: 16000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://hashicorp.com/careers",
	},
	{
		sourceID: "seed_004", title: "Software Engineer – Infrastructure", company: "PlanetScale",
		description: "PlanetScale is building the world's most advanced database platform, powered by Vitess and MySQL. Our infrastructure team keeps it all running reliably for thousands of customers.\n\nYou will:\n- Write Go services for our control plane\n- Work on database provisioning, backup, and restore automation\n- Improve observability with metrics and distributed tracing\n- Help design new platform features like multi-region branching\n\nStrong background in databases or distributed systems preferred.",
		skills:      []string{"go", "mysql", "kubernetes", "aws", "vitess"},
		salaryMin: 10000, salaryMax: 17000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://planetscale.com/careers",
	},
	{
		sourceID: "seed_005", title: "Staff Engineer – API Platform", company: "Temporal Technologies",
		description: "Temporal is the open-source workflow orchestration platform used by Stripe, Netflix, Coinbase, and thousands more. Join our API Platform team to shape the developer experience of the future.\n\nRole overview:\n- Lead architectural decisions for our Go SDK\n- Design backward-compatible API evolution strategies\n- Build developer tools and improve documentation\n- Work closely with customers to understand pain points\n\nWe are a remote-first company with a flat structure.",
		skills:      []string{"go", "grpc", "protobuf", "distributed systems", "postgres"},
		salaryMin: 14000, salaryMax: 24000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://temporal.io/careers",
	},
	{
		sourceID: "seed_006", title: "Backend Engineer", company: "Supabase",
		description: "Supabase is an open-source Firebase alternative built on Postgres. We're a small team building something used by hundreds of thousands of developers worldwide.\n\nYou'll work on:\n- Postgres extensions and tooling\n- Realtime subscription engine (Elixir + Go)\n- Auth service improvements\n- Storage and CDN infrastructure\n\nWe care about open source, developer experience, and shipping fast.",
		skills:      []string{"go", "postgres", "elixir", "docker", "aws"},
		salaryMin: 8000, salaryMax: 14000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://supabase.com/careers",
	},
	{
		sourceID: "seed_007", title: "Senior Backend Engineer", company: "Fly.io",
		description: "Fly.io is building a platform to deploy apps close to users, anywhere in the world. We run infrastructure for thousands of developers and are growing fast.\n\nThe role:\n- Work in Go on our orchestration layer and machine management API\n- Build reliable distributed systems on top of WireGuard\n- Improve our Anycast networking infrastructure\n- Debug gnarly networking problems\n\nSmall team, big impact. We move fast.",
		skills:      []string{"go", "distributed systems", "networking", "linux", "wireguard"},
		salaryMin: 9000, salaryMax: 16000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://fly.io/jobs",
	},
	{
		sourceID: "seed_008", title: "Backend Engineer – Billing", company: "Linear",
		description: "Linear is a project management tool loved by engineering teams around the world. We're a small, high-performing team and are looking for a backend engineer to own our billing and subscription systems.\n\nResponsibilities:\n- Integrate and maintain Stripe subscription billing\n- Build usage metering and quota enforcement systems\n- Design self-serve plan upgrade/downgrade flows\n- Ensure correctness with thorough testing\n\nStack: Node.js (TypeScript), Postgres, Redis, Go for some services.",
		skills:      []string{"go", "nodejs", "typescript", "postgres", "stripe"},
		salaryMin: 9000, salaryMax: 15000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://linear.app/careers",
	},
	{
		sourceID: "seed_009", title: "Go Developer – Security", company: "Tailscale",
		description: "Tailscale makes networking easy by building on WireGuard. Security is at the core of everything we do. Join our security engineering team to harden the platform and maintain our zero-trust model.\n\nYou'll work on:\n- Audit logging and compliance features\n- Certificate management and key rotation\n- Secure relay infrastructure\n- Fuzz testing and vulnerability research\n\nMust be comfortable reading and writing Go. Security background strongly preferred.",
		skills:      []string{"go", "wireguard", "networking", "linux", "cryptography"},
		salaryMin: 10000, salaryMax: 18000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://tailscale.com/jobs",
	},
	{
		sourceID: "seed_010", title: "Backend Engineer – Core", company: "Warp",
		description: "Warp is the intelligent terminal, built with Rust on the frontend and Go on the backend. We're building AI-powered features that make developers 10x more productive.\n\nBackend role:\n- Build REST and WebSocket APIs in Go\n- Integrate with Claude and GPT-4 APIs for AI features\n- Design secure multi-tenant data storage\n- Maintain our auth and permissions system\n\n3+ years Go experience required.",
		skills:      []string{"go", "postgres", "redis", "websockets", "aws"},
		salaryMin: 9500, salaryMax: 16000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://warp.dev/careers",
	},

	// ── Python / Data / ML ──────────────────────────────────────────────────
	{
		sourceID: "seed_011", title: "ML Engineer – LLM Fine-tuning", company: "Mistral AI",
		description: "Mistral AI is building the world's best open-source language models. Join our ML team to work on pre-training, fine-tuning, and RLHF pipelines.\n\nYou'll:\n- Run large-scale distributed training jobs on H100 clusters\n- Design data pipelines for instruction tuning datasets\n- Implement and evaluate alignment techniques (RLHF, DPO, PPO)\n- Benchmark models against MMLU, HumanEval, and internal evals\n\nPyTorch expert required. Experience with Megatron-LM or DeepSpeed is a big plus.",
		skills:      []string{"python", "pytorch", "cuda", "distributed training", "mlflow", "docker"},
		salaryMin: 12000, salaryMax: 22000, isRemote: false, location: "Paris, France",
		jobType: "full_time", url: "https://mistral.ai/careers",
	},
	{
		sourceID: "seed_012", title: "Senior ML Engineer – Recommendations", company: "Spotify",
		description: "Spotify's recommendation platform serves personalized music and podcasts to 600M+ users. We're looking for an ML engineer to improve our ranking and retrieval models.\n\nKey responsibilities:\n- Train and deploy deep learning models for collaborative filtering\n- A/B test recommendation algorithms at scale\n- Build ML pipelines with Airflow and Kubeflow\n- Collaborate with data scientists on feature engineering\n\nExperience with large-scale recommendation systems required.",
		skills:      []string{"python", "pytorch", "tensorflow", "airflow", "kafka", "sql", "spark"},
		salaryMin: 10000, salaryMax: 18000, isRemote: false, location: "Stockholm, Sweden",
		jobType: "full_time", url: "https://spotify.com/jobs",
	},
	{
		sourceID: "seed_013", title: "Data Engineer – Platform", company: "Databricks",
		description: "Databricks is the data + AI company. Our data platform team builds the pipelines that power everything from product analytics to ML training data.\n\nRole:\n- Build and maintain PB-scale data pipelines on Delta Lake\n- Design data quality monitoring and alerting\n- Create self-serve analytics infrastructure with dbt\n- Optimize Spark jobs for cost and performance\n\nDatabricks experience a plus. dbt and SQL expertise required.",
		skills:      []string{"python", "spark", "dbt", "sql", "kafka", "airflow", "delta lake"},
		salaryMin: 10000, salaryMax: 17000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://databricks.com/careers",
	},
	{
		sourceID: "seed_014", title: "Senior Data Engineer", company: "Notion",
		description: "Notion is the connected workspace used by millions. Our data team is responsible for building the infrastructure that enables data-driven decisions across the company.\n\nYou'll build:\n- Real-time and batch ingestion pipelines\n- Semantic layer with dbt and Looker\n- Data quality monitoring with Great Expectations\n- Self-serve analytics tooling for PMs and designers\n\nStrong SQL and Python required. Snowflake experience preferred.",
		skills:      []string{"python", "sql", "dbt", "snowflake", "airflow", "kafka"},
		salaryMin: 9000, salaryMax: 16000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://notion.so/careers",
	},
	{
		sourceID: "seed_015", title: "Python Backend Engineer – API", company: "Anthropic",
		description: "Anthropic is an AI safety company building reliable, interpretable AI systems. Our API team provides access to Claude to thousands of developers worldwide.\n\nYou'll work on:\n- FastAPI-based backend services at scale\n- API rate limiting, billing, and quota management\n- Model serving infrastructure and autoscaling\n- Developer portal and SDK improvements\n\nPython expert required. Experience with async Python and distributed systems preferred.",
		skills:      []string{"python", "fastapi", "postgres", "redis", "kubernetes", "aws"},
		salaryMin: 13000, salaryMax: 22000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://anthropic.com/careers",
	},
	{
		sourceID: "seed_016", title: "ML Infrastructure Engineer", company: "Hugging Face",
		description: "Hugging Face hosts 500,000+ models and datasets. Our infrastructure team keeps this all running and makes it easy for the community to share and run models.\n\nYou'll build:\n- Scalable model serving infrastructure (CPU and GPU)\n- Inference optimization pipelines (quantization, compilation)\n- Model Hub storage and CDN\n- Gradio Spaces compute backend\n\nStrong Python and ML systems experience required. CUDA knowledge a plus.",
		skills:      []string{"python", "pytorch", "cuda", "kubernetes", "docker", "aws"},
		salaryMin: 10000, salaryMax: 18000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://huggingface.co/jobs",
	},
	{
		sourceID: "seed_017", title: "Senior Python Developer – FinTech", company: "Revolut",
		description: "Revolut is building the world's first truly global financial superapp. Our engineering team processes millions of transactions per day across 35+ currencies.\n\nWe need a Python engineer to:\n- Build high-throughput payment processing services\n- Design fraud detection pipelines\n- Integrate with banking APIs (SWIFT, SEPA, ACH)\n- Maintain compliance and regulatory reporting\n\nPython expert required. Django or FastAPI experience. Previous fintech experience a big plus.",
		skills:      []string{"python", "django", "fastapi", "postgres", "kafka", "redis", "celery"},
		salaryMin: 7000, salaryMax: 12000, isRemote: true, location: "Europe",
		jobType: "full_time", url: "https://revolut.com/careers",
	},
	{
		sourceID: "seed_018", title: "Data Scientist – Growth", company: "Figma",
		description: "Figma is a design tool used by 4M+ designers. Our growth data science team uses data to drive product decisions and user acquisition strategy.\n\nYour work:\n- Build growth models and LTV predictions\n- Design and analyze A/B experiments\n- Create dashboards for key business metrics\n- Partner with marketing on attribution modeling\n\nStrong Python, SQL, and statistics background required. Experience with causal inference a plus.",
		skills:      []string{"python", "sql", "r", "spark", "airflow", "looker"},
		salaryMin: 9000, salaryMax: 16000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://figma.com/careers",
	},

	// ── Frontend / React ───────────────────────────────────────────────────────
	{
		sourceID: "seed_019", title: "Senior Frontend Engineer", company: "Vercel",
		description: "Vercel powers the frontend cloud. We're looking for a senior frontend engineer to work on our dashboard and developer experience.\n\nYou'll build:\n- Complex React applications with Next.js\n- Real-time deployment logs and monitoring UIs\n- Design system components with Radix UI and Tailwind\n- High-performance web experiences with Core Web Vitals in mind\n\nDeep React and Next.js expertise required. Performance optimization experience a must.",
		skills:      []string{"react", "nextjs", "typescript", "tailwind", "graphql"},
		salaryMin: 9000, salaryMax: 16000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://vercel.com/careers",
	},
	{
		sourceID: "seed_020", title: "Frontend Engineer – Design System", company: "Atlassian",
		description: "Atlassian's design system (Atlaskit) is used by thousands of developers building on top of Jira, Confluence, and Trello. Join our DS team to shape the frontend foundations.\n\nResponsibilities:\n- Build accessible, composable React components\n- Maintain Storybook documentation\n- Establish CSS-in-JS patterns and tokens\n- Collaborate with designers on component APIs\n\nStrong React and accessibility (a11y) knowledge required. Emotion or Styled Components experience preferred.",
		skills:      []string{"react", "typescript", "css", "storybook", "accessibility", "figma"},
		salaryMin: 8000, salaryMax: 14000, isRemote: true, location: "Australia, USA",
		jobType: "full_time", url: "https://atlassian.com/jobs",
	},
	{
		sourceID: "seed_021", title: "React Native Engineer", company: "Shopify",
		description: "Shopify's mobile apps are used by millions of merchants to run their businesses. Our React Native team is building the future of mobile commerce.\n\nYou'll work on:\n- Point of Sale (POS) features in React Native\n- Performance optimization for low-end Android devices\n- Native module bridges for payments and hardware\n- Shared component library with Polaris Mobile\n\n3+ years of React Native experience required.",
		skills:      []string{"react native", "typescript", "react", "ios", "android", "graphql"},
		salaryMin: 8000, salaryMax: 14000, isRemote: true, location: "Canada, USA, Europe",
		jobType: "full_time", url: "https://shopify.com/careers",
	},
	{
		sourceID: "seed_022", title: "Frontend Engineer – Editor", company: "Notion",
		description: "The Notion editor is a block-based rich text editor used by millions of people every day. It's one of the most complex pieces of software in our product.\n\nYou'll work on:\n- Block editor internals and plugin system\n- Real-time collaboration with CRDTs\n- Drag-and-drop and keyboard navigation\n- Performance optimization for large pages\n\nDeep JavaScript and browser expertise required. Prior work on editors (ProseMirror, Slate, Lexical) a strong plus.",
		skills:      []string{"typescript", "react", "prosemirror", "yjs", "websockets"},
		salaryMin: 10000, salaryMax: 18000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://notion.so/careers",
	},
	{
		sourceID: "seed_023", title: "Frontend Engineer – Data Visualization", company: "Grafana Labs",
		description: "Grafana is the most popular open-source observability platform. Our frontend team is building the data visualization capabilities that millions of engineers rely on.\n\nYou'll work on:\n- Time series visualization engine with Canvas and WebGL\n- New panel types and visualization primitives\n- Query editor improvements for PromQL, Loki, SQL\n- Grafana plugin architecture and SDK\n\nReact and TypeScript required. D3.js or Canvas API experience a plus.",
		skills:      []string{"typescript", "react", "d3", "canvas", "grafana"},
		salaryMin: 8000, salaryMax: 14000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://grafana.com/careers",
	},
	{
		sourceID: "seed_024", title: "Frontend Engineer", company: "Cursor",
		description: "Cursor is an AI-first code editor built on VS Code. We're looking for a frontend engineer who cares deeply about developer tools and editing experiences.\n\nYou'll work on:\n- AI features: chat, autocomplete, code edit suggestions\n- Editor performance and startup time\n- Extension system and plugin APIs\n- UI/UX improvements to the Cursor interface\n\nElectron and VS Code extension API experience helpful but not required. Strong TypeScript required.",
		skills:      []string{"typescript", "react", "electron", "vscode api", "nodejs"},
		salaryMin: 10000, salaryMax: 18000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://cursor.com/careers",
	},

	// ── DevOps / SRE / Platform ───────────────────────────────────────────────
	{
		sourceID: "seed_025", title: "Senior SRE", company: "Google",
		description: "Google SRE invented the SRE discipline. Join our team to ensure reliability for products used by billions of people.\n\nResponsibilities:\n- Own SLOs and error budget policy for critical services\n- Reduce toil through automation and improved tooling\n- Lead incident response and postmortems\n- Design for reliability at billion-user scale\n\nStrong Linux, networking, and programming (Python or Go) background required.",
		skills:      []string{"go", "python", "kubernetes", "linux", "prometheus", "terraform"},
		salaryMin: 15000, salaryMax: 25000, isRemote: false, location: "Mountain View, USA",
		jobType: "full_time", url: "https://careers.google.com",
	},
	{
		sourceID: "seed_026", title: "DevOps Engineer", company: "GitLab",
		description: "GitLab is an all-remote company with 2,000+ team members in 65 countries. Our infrastructure team manages the platform that millions of developers use every day.\n\nYou'll work on:\n- Kubernetes cluster management and upgrades\n- CI/CD pipeline infrastructure for GitLab.com\n- Autoscaling and cost optimization\n- Disaster recovery and backup systems\n\nTerraform and Kubernetes expertise required. GCP experience preferred.",
		skills:      []string{"kubernetes", "terraform", "gcp", "prometheus", "grafana", "python", "ruby"},
		salaryMin: 8000, salaryMax: 14000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://about.gitlab.com/jobs",
	},
	{
		sourceID: "seed_027", title: "Platform Engineer", company: "Shopify",
		description: "Shopify's platform engineering team builds the internal tools and infrastructure that 10,000+ engineers rely on every day.\n\nYou'll build:\n- Developer portal and service catalog\n- Internal CI/CD tooling and release automation\n- Service mesh and API gateway configuration\n- Cost visibility and FinOps tooling\n\nStrong Go and Kubernetes knowledge required. Experience with large-scale engineering platforms a plus.",
		skills:      []string{"go", "kubernetes", "terraform", "aws", "grpc", "postgres"},
		salaryMin: 10000, salaryMax: 18000, isRemote: true, location: "Canada, USA",
		jobType: "full_time", url: "https://shopify.com/careers",
	},
	{
		sourceID: "seed_028", title: "Infrastructure Engineer – Compute", company: "Railway",
		description: "Railway is a deployment platform that makes it easy for developers to ship their apps. Our compute team runs thousands of containers for customers around the world.\n\nYou'll work on:\n- Container orchestration and scheduling on bare metal\n- Build and run systems (Nixpacks-based builds)\n- Networking and overlay networks with WireGuard\n- GPU workload scheduling for ML workloads\n\nSmall team, high ownership, significant impact.",
		skills:      []string{"go", "linux", "kubernetes", "wireguard", "rust", "networking"},
		salaryMin: 9000, salaryMax: 16000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://railway.app/careers",
	},
	{
		sourceID: "seed_029", title: "Cloud Infrastructure Engineer", company: "Render",
		description: "Render is a cloud platform that's replacing Heroku for thousands of developers. Our infrastructure team is responsible for the reliability and performance of our multi-region cloud.\n\nRole:\n- Manage AWS infrastructure with Terraform\n- Build autoscaling systems for web services and databases\n- Improve container startup time and efficiency\n- Implement multi-region failover\n\nAWS and Terraform expertise required. Go experience a plus.",
		skills:      []string{"aws", "terraform", "kubernetes", "go", "postgres", "redis"},
		salaryMin: 8500, salaryMax: 15000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://render.com/jobs",
	},

	// ── Rust ─────────────────────────────────────────────────────────────────
	{
		sourceID: "seed_030", title: "Systems Engineer – Rust", company: "Cloudflare",
		description: "Cloudflare is building more and more of its edge software in Rust. Join our systems team to write safe, fast, and correct code that runs at the edge in 250+ cities.\n\nYou'll work on:\n- Proxy internals and HTTP processing pipeline\n- WebAssembly runtime for Workers\n- eBPF programs for packet filtering\n- Memory-safe reimplementations of critical C components\n\nDeep Rust expertise required. Low-level systems programming background essential.",
		skills:      []string{"rust", "c", "wasm", "linux", "ebpf", "networking"},
		salaryMin: 12000, salaryMax: 20000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://cloudflare.com/careers",
	},
	{
		sourceID: "seed_031", title: "Rust Developer – Database Engine", company: "TigerBeetle",
		description: "TigerBeetle is building a financial accounting database that can process 1M+ transactions per second. Every line of code matters.\n\nYou'll work directly on:\n- The core database engine written in Zig and Rust\n- LSMT storage engine and WAL implementation\n- Distributed consensus with Viewstamped Replication\n- Deterministic simulation testing\n\nExtreme attention to correctness required. Experience with storage engines or databases a strong plus.",
		skills:      []string{"rust", "zig", "distributed systems", "databases", "linux"},
		salaryMin: 12000, salaryMax: 22000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://tigerbeetle.com/careers",
	},
	{
		sourceID: "seed_032", title: "Rust Engineer – CLI Tooling", company: "Warp",
		description: "Warp is a Rust-based terminal application. Our desktop team works on the core editing experience, AI integration, and native platform features.\n\nYou'll build:\n- Terminal emulator internals (VTE parsing, rendering)\n- GPU-accelerated text rendering with wgpu\n- AI-powered shell completions and command search\n- Cross-platform native features (macOS, Linux, Windows)\n\nStrong Rust experience required. GUI programming experience a plus.",
		skills:      []string{"rust", "wgpu", "tokio", "macos", "linux", "wasm"},
		salaryMin: 10000, salaryMax: 18000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://warp.dev/careers",
	},

	// ── Java / JVM ────────────────────────────────────────────────────────────
	{
		sourceID: "seed_033", title: "Java Backend Engineer", company: "Netflix",
		description: "Netflix streams to 260M+ subscribers using a massive microservices architecture built largely in Java. Our platform team is hiring a backend engineer to work on our streaming infrastructure.\n\nResponsibilities:\n- Build and maintain Java Spring Boot microservices\n- Improve our service mesh and API gateway (Zuul/Envoy)\n- Work on video ingestion and transcoding pipelines\n- Participate in chaos engineering initiatives\n\n5+ years of Java experience required. Experience with large-scale distributed systems a must.",
		skills:      []string{"java", "spring", "kafka", "cassandra", "aws", "docker", "kubernetes"},
		salaryMin: 12000, salaryMax: 20000, isRemote: false, location: "Los Gatos, USA",
		jobType: "full_time", url: "https://jobs.netflix.com",
	},
	{
		sourceID: "seed_034", title: "Senior Kotlin Developer – Backend", company: "JetBrains",
		description: "JetBrains creates Kotlin and IntelliJ IDEA. Join our backend team to build the services that power JetBrains Marketplace, License Server, and our team tools.\n\nYou'll work on:\n- Ktor-based microservices\n- Payment and licensing systems\n- Developer-facing APIs and webhooks\n- Internal tooling with Exposed ORM\n\nStrong Kotlin and JVM experience required. Coroutines expertise expected.",
		skills:      []string{"kotlin", "java", "ktor", "postgres", "redis", "docker"},
		salaryMin: 7000, salaryMax: 12000, isRemote: true, location: "Europe",
		jobType: "full_time", url: "https://jetbrains.com/careers",
	},
	{
		sourceID: "seed_035", title: "Backend Engineer – Scala", company: "Twitter / X",
		description: "Twitter's backend is built on Scala and the Finagle RPC framework created internally. Join our core platform team to work on systems serving 250M+ daily users.\n\nYou'll work on:\n- Core timeline ranking and delivery services\n- Real-time event processing with Apache Flink\n- Twitter's home-grown Scala microservices framework\n- Improving reliability and reducing latency\n\nScala or functional programming experience required.",
		skills:      []string{"scala", "java", "kafka", "hadoop", "spark", "mysql"},
		salaryMin: 10000, salaryMax: 18000, isRemote: false, location: "San Francisco, USA",
		jobType: "full_time", url: "https://careers.twitter.com",
	},

	// ── iOS / Mobile ─────────────────────────────────────────────────────────
	{
		sourceID: "seed_036", title: "iOS Engineer – Core", company: "Airbnb",
		description: "Airbnb's iOS app is used by millions of hosts and guests worldwide. Our core iOS team is responsible for the app's architecture, performance, and developer experience.\n\nYou'll work on:\n- Airbnb's proprietary iOS architecture (Epoxy)\n- App performance profiling and optimization\n- Internal Swift tooling and code generation\n- Migration from UIKit to SwiftUI\n\n4+ years of iOS experience required. SwiftUI expertise a plus.",
		skills:      []string{"swift", "swiftui", "ios", "xcode", "objective-c"},
		salaryMin: 10000, salaryMax: 18000, isRemote: false, location: "San Francisco, USA",
		jobType: "full_time", url: "https://careers.airbnb.com",
	},
	{
		sourceID: "seed_037", title: "iOS Developer", company: "Spotify",
		description: "Spotify's iOS app streams music and podcasts to hundreds of millions of users. Join our mobile team to build features that affect music lovers worldwide.\n\nResponsibilities:\n- Implement new features in Swift and SwiftUI\n- A/B testing with our internal experimentation framework\n- Improve streaming reliability and offline mode\n- Contribute to our open-source iOS libraries\n\n3+ years of Swift experience required.",
		skills:      []string{"swift", "swiftui", "ios", "xcode", "objective-c", "ci/cd"},
		salaryMin: 8000, salaryMax: 15000, isRemote: false, location: "Stockholm, Sweden",
		jobType: "full_time", url: "https://spotify.com/jobs",
	},
	{
		sourceID: "seed_038", title: "Senior iOS Engineer", company: "Duolingo",
		description: "Duolingo is the most popular language learning app in the world with 80M+ daily active users. Our iOS team builds and optimizes the gamified learning experience.\n\nYou'll work on:\n- Core lesson engine and gamification features\n- App performance (startup, frame rate, memory)\n- Accessibility improvements\n- Streaks, leaderboards, and social features\n\nSwift and UIKit expertise required. SwiftUI experience a plus.",
		skills:      []string{"swift", "ios", "swiftui", "xcode", "animations"},
		salaryMin: 9000, salaryMax: 16000, isRemote: false, location: "Pittsburgh, USA",
		jobType: "full_time", url: "https://duolingo.com/careers",
	},

	// ── Android ───────────────────────────────────────────────────────────────
	{
		sourceID: "seed_039", title: "Android Engineer – Core", company: "WhatsApp",
		description: "WhatsApp serves 2B+ users worldwide. Our Android team is responsible for one of the most-used apps in the world. Join us to build features that matter at global scale.\n\nYou'll work on:\n- Core messaging infrastructure\n- End-to-end encryption implementation\n- Voice and video calling features\n- App performance and battery optimization\n\nStrong Kotlin and Android expertise required. Java background helpful.",
		skills:      []string{"kotlin", "android", "java", "c++", "jni"},
		salaryMin: 12000, salaryMax: 20000, isRemote: false, location: "Menlo Park, USA",
		jobType: "full_time", url: "https://careers.whatsapp.com",
	},
	{
		sourceID: "seed_040", title: "Android Developer", company: "N26",
		description: "N26 is a mobile bank with 8M+ customers across Europe and the US. Our Android team is building the future of mobile banking.\n\nRole:\n- Implement new banking features in Kotlin\n- Maintain our design system components with Jetpack Compose\n- Work on biometric authentication and security features\n- Performance and accessibility improvements\n\n3+ years Kotlin/Android required. Fintech or security background a plus.",
		skills:      []string{"kotlin", "android", "jetpack compose", "coroutines", "mvvm"},
		salaryMin: 6500, salaryMax: 11000, isRemote: false, location: "Berlin, Germany",
		jobType: "full_time", url: "https://n26.com/jobs",
	},

	// ── Full-stack / Node.js ─────────────────────────────────────────────────
	{
		sourceID: "seed_041", title: "Full-stack Engineer", company: "Replit",
		description: "Replit is a collaborative browser-based IDE used by 30M+ developers and students. Our team builds the entire product: editor, runtime, multiplayer, AI features.\n\nYou'll work across the stack:\n- React + TypeScript frontend for the IDE\n- Node.js backend services\n- Docker-based container execution\n- AI coding assistant features\n\nFull-stack generalist who can ship end-to-end features required.",
		skills:      []string{"typescript", "react", "nodejs", "postgres", "redis", "docker"},
		salaryMin: 9000, salaryMax: 16000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://replit.com/careers",
	},
	{
		sourceID: "seed_042", title: "Node.js Backend Engineer", company: "Twilio",
		description: "Twilio is a cloud communications platform used by 300,000+ companies. Our messaging team processes billions of SMS, voice, and email messages per month.\n\nYou'll work on:\n- High-throughput Node.js message processing services\n- Carrier integrations and webhook delivery\n- SMS filtering and spam prevention\n- Global phone number management APIs\n\nNode.js and async programming expertise required.",
		skills:      []string{"nodejs", "typescript", "postgres", "redis", "kafka", "aws"},
		salaryMin: 8000, salaryMax: 14000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://twilio.com/jobs",
	},
	{
		sourceID: "seed_043", title: "Full-stack Engineer – SaaS Platform", company: "Intercom",
		description: "Intercom is the AI-first customer service platform. We serve 25,000+ businesses and process millions of customer conversations. Join our platform team to build the infrastructure for scale.\n\nYou'll work on:\n- Ruby on Rails backend (migrating to Node.js)\n- React frontend components\n- Real-time messaging with WebSockets\n- Elasticsearch integration for conversation search\n\nStrong full-stack experience required. Ruby or Node.js backend, React frontend.",
		skills:      []string{"ruby", "nodejs", "typescript", "react", "postgres", "elasticsearch", "redis"},
		salaryMin: 9000, salaryMax: 16000, isRemote: true, location: "Ireland, USA",
		jobType: "full_time", url: "https://intercom.com/careers",
	},

	// ── Contract / Part-time ─────────────────────────────────────────────────
	{
		sourceID: "seed_044", title: "Go Developer – Contract", company: "Freelance / Remote",
		description: "We're looking for an experienced Go developer for a 3-month contract to help build a microservices migration from a Python monolith. The work involves designing gRPC services, writing comprehensive tests, and documenting architecture decisions.\n\nExpected hours: 40h/week for 3 months. Fully remote. Possibility of extension or full-time offer.",
		skills:      []string{"go", "grpc", "postgres", "docker", "microservices"},
		salaryMin: 6000, salaryMax: 10000, isRemote: true, location: "Worldwide",
		jobType: "contract", url: "https://toptal.com",
	},
	{
		sourceID: "seed_045", title: "Part-time React Developer", company: "Design Agency Berlin",
		description: "Small design agency looking for a part-time React developer (20h/week) to build interactive web experiences for our clients. Projects include e-commerce sites, landing pages, and custom dashboards.\n\nFlexible hours. Preferred: Central European timezone. Rate: €50-80/hour.",
		skills:      []string{"react", "typescript", "css", "nextjs", "figma"},
		salaryMin: 2500, salaryMax: 4500, isRemote: true, location: "Germany (preferred)",
		jobType: "part_time", url: "https://example.com",
	},
	{
		sourceID: "seed_046", title: "Python Freelancer – Data Pipelines", company: "Analytics Startup",
		description: "Early-stage analytics startup seeking a Python freelancer to help build our data ingestion and transformation pipelines. You'll work directly with the CTO on architecture decisions.\n\nExpected scope: 60-80 hours over 6 weeks. Potential for ongoing engagement.",
		skills:      []string{"python", "dbt", "airflow", "sql", "postgres"},
		salaryMin: 3500, salaryMax: 6000, isRemote: true, location: "Worldwide",
		jobType: "contract", url: "https://upwork.com",
	},

	// ── Security / Crypto ─────────────────────────────────────────────────────
	{
		sourceID: "seed_047", title: "Security Engineer – Application", company: "GitHub",
		description: "GitHub is the world's leading software development platform. Our application security team protects the code of 100M+ developers.\n\nYou'll work on:\n- Code scanning and secret detection features\n- Vulnerability disclosure and bug bounty program\n- Penetration testing of GitHub.com\n- Security code review for new features\n\nStrong application security background required. Web exploitation and Ruby/Go experience a plus.",
		skills:      []string{"security", "ruby", "go", "python", "linux", "networking"},
		salaryMin: 11000, salaryMax: 20000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://github.com/about/careers",
	},
	{
		sourceID: "seed_048", title: "Cryptography Engineer", company: "1Password",
		description: "1Password protects the passwords and secrets of millions of people and businesses. Our cryptography team is responsible for our security architecture and encryption implementation.\n\nYou'll work on:\n- End-to-end encrypted vault design\n- Key derivation and authentication protocols\n- Secure multi-party computation research\n- Cryptographic library maintenance\n\nStrong background in applied cryptography required.",
		skills:      []string{"rust", "go", "cryptography", "security", "c++"},
		salaryMin: 12000, salaryMax: 20000, isRemote: true, location: "Canada, USA",
		jobType: "full_time", url: "https://1password.com/careers",
	},

	// ── QA / Testing ──────────────────────────────────────────────────────────
	{
		sourceID: "seed_049", title: "Senior QA Engineer – Automation", company: "Booking.com",
		description: "Booking.com processes millions of reservations per day. Our QA team ensures quality across web, mobile, and backend systems.\n\nYou'll build:\n- Selenium and Playwright-based UI test suites\n- API test automation frameworks in Python\n- Load and performance testing with Locust\n- Test infrastructure on Kubernetes\n\nStrong Python and test automation experience required.",
		skills:      []string{"python", "selenium", "playwright", "pytest", "kubernetes"},
		salaryMin: 6000, salaryMax: 10000, isRemote: false, location: "Amsterdam, Netherlands",
		jobType: "full_time", url: "https://booking.com/careers",
	},
	{
		sourceID: "seed_050", title: "QA Engineer – Mobile", company: "Duolingo",
		description: "Duolingo's QA team ensures that learning a language on mobile is a smooth and reliable experience for 80M+ daily users.\n\nResponsibilities:\n- Manual and automated testing for iOS and Android\n- Maintain Appium and XCTest automation suites\n- Coordinate release QA process\n- Perform regression and exploratory testing\n\n2+ years of mobile QA experience required.",
		skills:      []string{"appium", "selenium", "python", "ios", "android", "xcode"},
		salaryMin: 5000, salaryMax: 8500, isRemote: false, location: "Pittsburgh, USA",
		jobType: "full_time", url: "https://duolingo.com/careers",
	},

	// ── Additional roles ─────────────────────────────────────────────────────
	{
		sourceID: "seed_051", title: "Backend Engineer – Real-time", company: "Discord",
		description: "Discord serves 500M+ registered users with real-time messaging, voice, and video. Our backend team works on the systems that make this possible at scale.\n\nYou'll work on:\n- WebSocket gateway serving millions of concurrent connections\n- Message delivery and ordering guarantees\n- Presence and activity systems\n- Elixir-based services with Erlang/OTP\n\nElixir or Erlang experience preferred. Go or Rust also considered.",
		skills:      []string{"elixir", "go", "rust", "postgres", "redis", "cassandra"},
		salaryMin: 10000, salaryMax: 18000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://discord.com/careers",
	},
	{
		sourceID: "seed_052", title: "Staff Software Engineer – Search", company: "Elastic",
		description: "Elastic is the company behind Elasticsearch, Kibana, and Beats. Our search team works on the core query engine and ranking algorithms that power thousands of search applications.\n\nYou'll work on:\n- Query execution and optimization in Java\n- Approximate nearest neighbor (ANN) search for vector databases\n- Relevance ranking improvements\n- Search infrastructure on Kubernetes\n\nJava and information retrieval background required.",
		skills:      []string{"java", "elasticsearch", "lucene", "python", "docker", "kubernetes"},
		salaryMin: 13000, salaryMax: 22000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://elastic.co/careers",
	},
	{
		sourceID: "seed_053", title: "Embedded Systems Engineer", company: "Arduino",
		description: "Arduino makes electronics accessible to everyone. Our firmware team writes the software that runs on millions of microcontrollers worldwide.\n\nYou'll work on:\n- C/C++ firmware for ARM Cortex-M microcontrollers\n- Arduino Core libraries and BSPs\n- USB, BLE, and WiFi stack integration\n- Testing frameworks for embedded code\n\nStrong C/C++ and embedded systems background required.",
		skills:      []string{"c", "c++", "embedded systems", "arm", "ble", "linux"},
		salaryMin: 5000, salaryMax: 9000, isRemote: true, location: "Italy, Worldwide",
		jobType: "full_time", url: "https://arduino.cc/jobs",
	},
	{
		sourceID: "seed_054", title: "Backend Engineer – Marketplace", company: "Airbnb",
		description: "Airbnb's Marketplace team is responsible for the algorithms and infrastructure that match hosts and guests. We process billions of search queries and optimize for both sides of the marketplace.\n\nYou'll work on:\n- Search ranking and recommendation systems\n- Dynamic pricing algorithms\n- Supply and demand forecasting\n- A/B testing infrastructure\n\nJava or Python backend experience required. ML experience a plus.",
		skills:      []string{"java", "python", "spark", "kafka", "mysql", "redis", "aws"},
		salaryMin: 12000, salaryMax: 20000, isRemote: false, location: "San Francisco, USA",
		jobType: "full_time", url: "https://careers.airbnb.com",
	},
	{
		sourceID: "seed_055", title: "Go Developer – Open Source", company: "CNCF / Remote",
		description: "Work on open-source cloud-native projects under the Cloud Native Computing Foundation. This role focuses on contributing to projects like containerd, Prometheus, or Argo.\n\nYou'll spend 80% of your time:\n- Writing Go code for CNCF projects\n- Reviewing community pull requests\n- Writing documentation and blog posts\n- Attending KubeCon and contributor summits\n\nActive open-source contributor preferred.",
		skills:      []string{"go", "kubernetes", "docker", "prometheus", "linux"},
		salaryMin: 8000, salaryMax: 14000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://cncf.io",
	},
	{
		sourceID: "seed_056", title: "Blockchain Engineer – Rust", company: "Solana Labs",
		description: "Solana is one of the fastest blockchains in existence, processing 65,000+ TPS. Our core team writes Rust and C for the validator, runtime, and programs.\n\nYou'll work on:\n- Solana validator node software\n- BPF VM and program runtime\n- Networking and consensus (Tower BFT)\n- Performance benchmarking and optimization\n\nDeep Rust expertise required. Distributed systems or databases background preferred.",
		skills:      []string{"rust", "c", "distributed systems", "linux", "networking"},
		salaryMin: 12000, salaryMax: 22000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://solanalabs.com/careers",
	},
	{
		sourceID: "seed_057", title: "Senior Python Engineer – Developer Tools", company: "Sentry",
		description: "Sentry helps 90,000+ companies monitor their software in production. Our developer tools team builds the SDKs and integrations that developers install in their apps.\n\nYou'll work on:\n- Python SDK and its integrations (Django, FastAPI, Flask)\n- Error grouping and fingerprinting algorithms\n- Performance monitoring and profiling\n- Relay (our data pipeline written in Rust)\n\nStrong Python expertise required. Open-source experience a plus.",
		skills:      []string{"python", "django", "fastapi", "postgres", "redis", "kafka", "rust"},
		salaryMin: 8000, salaryMax: 14000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://sentry.io/careers",
	},
	{
		sourceID: "seed_058", title: "TypeScript Engineer – Tooling", company: "Vercel",
		description: "Vercel's Developer Experience team builds the tools that make the platform a joy to use: CLI, VS Code extension, and Next.js developer experience.\n\nYou'll work on:\n- Vercel CLI (Node.js + TypeScript)\n- Build system integrations (Turborepo, Next.js)\n- VS Code extension\n- Developer documentation and code examples\n\nStrong TypeScript and Node.js experience required.",
		skills:      []string{"typescript", "nodejs", "nextjs", "react", "cli"},
		salaryMin: 9000, salaryMax: 16000, isRemote: true, location: "Worldwide",
		jobType: "full_time", url: "https://vercel.com/careers",
	},
	{
		sourceID: "seed_059", title: "DevOps Engineer – ML Infrastructure", company: "Scale AI",
		description: "Scale AI powers AI development for leading AI labs and enterprise companies. Our ML infrastructure team runs the compute needed to label billions of data points and train models.\n\nYou'll work on:\n- GPU cluster management on AWS and GCP\n- Kubeflow-based ML training pipelines\n- Data storage and processing infrastructure\n- Cost optimization across petabytes of training data\n\nKubernetes and cloud infrastructure expertise required. ML pipeline experience a plus.",
		skills:      []string{"kubernetes", "python", "terraform", "aws", "gcp", "docker"},
		salaryMin: 10000, salaryMax: 18000, isRemote: true, location: "USA",
		jobType: "full_time", url: "https://scale.com/careers",
	},
	{
		sourceID: "seed_060", title: "Backend Engineer – Search", company: "Algolia",
		description: "Algolia powers search for 17,000+ companies including Stripe, Twitch, and Medium. Our backend team builds the distributed search engine at the core of our product.\n\nYou'll work on:\n- C++ search engine core (indexing, ranking, faceting)\n- Go-based API layer and control plane\n- Multi-region replication and consistency\n- Advanced features: personalization, A/B testing, NLP\n\nStrong distributed systems and search experience required.",
		skills:      []string{"go", "c++", "distributed systems", "elasticsearch", "postgres"},
		salaryMin: 9000, salaryMax: 16000, isRemote: true, location: "France, USA",
		jobType: "full_time", url: "https://algolia.com/careers",
	},
}

func seedJobs(ctx context.Context, db *pgxpool.Pool) []uuid.UUID {
	var ids []uuid.UUID
	for _, j := range jobs {
		var existing uuid.UUID
		_ = db.QueryRow(ctx, `SELECT id FROM jobs WHERE source_id = $1`, j.sourceID).Scan(&existing)
		if existing != uuid.Nil {
			ids = append(ids, existing)
			continue
		}

		id := uuid.New()
		_, err := db.Exec(ctx,
			`INSERT INTO jobs (id, title, company, description, skills, salary_min, salary_max, is_remote, location, job_type, source, source_id, url, is_active)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'seed',$11,$12,true)`,
			id, j.title, j.company, j.description, j.skills,
			j.salaryMin, j.salaryMax, j.isRemote, j.location, j.jobType,
			j.sourceID, j.url,
		)
		if err != nil {
			log.Printf("job %s: %v", j.sourceID, err)
			continue
		}
		ids = append(ids, id)
		fmt.Printf("  + %s @ %s\n", j.title, j.company)
	}
	return ids
}

// ─── applications ─────────────────────────────────────────────────────────────

type appSeed struct {
	userIdx int
	jobIdx  int
	status  string
	note    string
}

func seedApplications(ctx context.Context, db *pgxpool.Pool, userIDs, jobIDs []uuid.UUID, rng *rand.Rand) {
	statuses := []string{"saved", "applied", "interview", "offer", "rejected"}

	applications := []appSeed{
		{1, 0, "interview", "Great match for my Go skills. Cloudflare is dream company."},
		{1, 2, "applied", "HashiCorp has great remote culture."},
		{1, 3, "saved", "Interested in PlanetScale's Vitess-based platform."},
		{1, 6, "applied", "Fly.io is exciting. Small team, big infra challenges."},
		{1, 8, "rejected", "Tailscale reached out but ghosted after first screen."},
		{1, 54, "saved", "Open source role at CNCF looks interesting."},

		{2, 10, "offer", "Mistral offered €18k. Negotiating start date."},
		{2, 12, "interview", "Databricks third round, very positive."},
		{2, 14, "applied", "Anthropic is where I want to be."},
		{2, 15, "saved", "HuggingFace always on my radar."},
		{2, 57, "applied", "Sentry uses Python heavily, good fit."},

		{3, 18, "applied", "Vercel is perfect for my Next.js experience."},
		{3, 19, "interview", "Atlassian DS team looks amazing."},
		{3, 21, "saved", "Notion editor challenge is exciting."},
		{3, 22, "applied", "Grafana data viz role, unique opportunity."},
		{3, 23, "rejected", "Cursor turned me down, skills gap in Electron."},
		{3, 57, "saved", ""},

		{4, 24, "applied", "Google SRE is the pinnacle."},
		{4, 25, "interview", "GitLab is fully remote which I love."},
		{4, 26, "applied", "Shopify platform eng, great opportunity."},
		{4, 28, "saved", "Render cloud infra looks interesting."},
		{4, 58, "applied", "Scale AI ML infra, very relevant."},

		{5, 29, "interview", "Cloudflare Rust team! Dream role."},
		{5, 30, "applied", "TigerBeetle is hardcore engineering."},
		{5, 31, "saved", "Warp terminal is awesome product."},
		{5, 8, "applied", "Tailscale security + Go + Rust, perfect match."},

		{6, 7, "applied", "Linear seems like a great team to work with."},
		{6, 40, "interview", "Replit full-stack, love their product."},
		{6, 41, "applied", "Twilio Node.js role, direct experience."},
		{6, 42, "saved", "Intercom is interesting, Ruby to Node migration."},

		{7, 11, "applied", "Spotify data eng, perfect for my skills."},
		{7, 12, "interview", "Databricks is data engineering heaven."},
		{7, 13, "offer", "Notion offered $12k/month. Evaluating."},
		{7, 16, "saved", "Revolut fintech, always wanted to work in payments."},

		{8, 32, "saved", "Netflix Java role, very senior level."},
		{8, 33, "applied", "JetBrains Kotlin team! Working on Kotlin itself."},
		{8, 51, "applied", "Elastic search team, relevant Java experience."},

		{9, 35, "applied", "Airbnb iOS core team, prestigious."},
		{9, 36, "interview", "Spotify iOS, great product."},
		{9, 37, "saved", "Duolingo is fun company to work at."},

		{10, 38, "interview", "WhatsApp Android, massive scale."},
		{10, 39, "applied", "N26 mobile banking in Berlin, like the city."},
		{10, 40, "saved", "Replit React Native, interesting mix."},

		{11, 16, "applied", "Revolut Python team, fintech experience."},
		{11, 14, "interview", "Anthropic FastAPI role, AI company."},
		{11, 55, "saved", "Sentry Python SDK, love open source."},
		{11, 46, "applied", "Security at GitHub is top of game."},
	}

	inserted := 0
	for _, a := range applications {
		if a.userIdx >= len(userIDs) || a.jobIdx >= len(jobIDs) {
			continue
		}
		userID := userIDs[a.userIdx]
		jobID := jobIDs[a.jobIdx]
		if userID == uuid.Nil || jobID == uuid.Nil {
			continue
		}

		var appliedAt *time.Time
		if a.status == "applied" || a.status == "interview" || a.status == "offer" || a.status == "rejected" {
			t := time.Now().Add(-time.Duration(rng.Intn(30)+1) * 24 * time.Hour)
			appliedAt = &t
		}

		_, err := db.Exec(ctx,
			`INSERT INTO applications (id, user_id, job_id, status, note, applied_at)
			 VALUES ($1,$2,$3,$4,$5,$6)
			 ON CONFLICT (user_id, job_id) DO NOTHING`,
			uuid.New(), userID, jobID, a.status, a.note, appliedAt,
		)
		if err != nil {
			log.Printf("application user=%s job=%s: %v", userID, jobID, err)
			continue
		}
		_ = statuses
		inserted++
	}
	fmt.Printf("  inserted %d applications\n", inserted)
}

// ─── saved jobs ───────────────────────────────────────────────────────────────

func seedSavedJobs(ctx context.Context, db *pgxpool.Pool, userIDs, jobIDs []uuid.UUID, rng *rand.Rand) {
	pairs := [][2]int{
		{1, 4}, {1, 9}, {1, 54}, {1, 55},
		{2, 11}, {2, 13}, {2, 15}, {2, 16},
		{3, 18}, {3, 20}, {3, 22}, {3, 23},
		{4, 25}, {4, 27}, {4, 28}, {4, 58},
		{5, 29}, {5, 30}, {5, 47},
		{6, 41}, {6, 42},
		{7, 12}, {7, 17},
		{8, 34}, {8, 51},
		{9, 36}, {9, 37},
		{10, 38}, {10, 39},
		{11, 15}, {11, 55},
	}

	inserted := 0
	for _, p := range pairs {
		ui, ji := p[0], p[1]
		if ui >= len(userIDs) || ji >= len(jobIDs) {
			continue
		}
		_, err := db.Exec(ctx,
			`INSERT INTO saved_jobs (id, user_id, job_id) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
			uuid.New(), userIDs[ui], jobIDs[ji],
		)
		if err != nil {
			continue
		}
		inserted++
	}
	_ = rng
	fmt.Printf("  inserted %d saved jobs\n", inserted)
}
