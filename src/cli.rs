use clap::{Args, Parser, Subcommand, ValueEnum};

#[derive(Parser)]
#[command(
    name = "clikd",
    about = "Development CLI for Clikd platform",
    long_about = "A powerful CLI tool for managing your Clikd development environment.\nProvides commands for authentication, service orchestration, and container monitoring.",
    version,
    after_help = "For detailed command help, run: clikd <COMMAND> --help"
)]
#[command(disable_version_flag = true)]
pub struct Cli {
    #[arg(
        short,
        long,
        action = clap::ArgAction::Count,
        global = true,
        help = "Increase logging verbosity (-v, -vv, -vvv)"
    )]
    pub verbose: u8,

    #[arg(long, global = true, help = "Disable colored output")]
    pub no_color: bool,

    #[arg(
        short,
        long,
        global = true,
        env = "CLIKD_ENV",
        help = "Environment configuration to use (development, staging, production)"
    )]
    pub env: Option<String>,

    #[arg(short = 'V', long, help = "Print version information")]
    pub version: bool,

    #[command(subcommand)]
    pub command: Option<Commands>,
}

#[derive(Subcommand)]
pub enum Commands {
    #[command(about = "Authenticate with Clikd platform")]
    Login {
        #[arg(long, help = "Skip opening browser and show URL only")]
        no_browser: bool,
    },

    #[command(about = "Sign out from Clikd platform")]
    Logout,

    #[command(subcommand, about = "Authentication management commands")]
    Auth(AuthCommands),

    #[command(about = "Initialize a new Clikd project")]
    Init(InitArgs),

    #[command(
        about = "Start all services",
        long_about = "Starts all configured services in Docker containers.\nCreates network, pulls images, and ensures health checks pass."
    )]
    Start(StartArgs),

    #[command(
        about = "Stop all running services",
        long_about = "Stops all running containers.\nUse --purge to also remove volumes and clean up completely."
    )]
    Stop(StopArgs),

    #[command(
        about = "Interactive container monitoring TUI",
        long_about = "Launch an interactive terminal UI for real-time container monitoring.\n\nFeatures:\n  • Live container metrics (CPU, memory, network)\n  • Interactive log viewer with search and export\n  • Container controls (start, stop, restart, pause, delete)\n  • Sortable columns and mouse support\n  • Press 'h' in TUI for keyboard shortcuts"
    )]
    Status(StatusArgs),

    #[command(about = "Update CLI to the latest version")]
    Update(UpdateArgs),

    #[command(about = "Generate shell completions")]
    Completions {
        #[arg(value_enum, help = "Shell type to generate completions for")]
        shell: clap_complete::Shell,
    },

    #[command(
        subcommand,
        about = "Release management commands",
        long_about = "Powerful release management for monorepos and multi-language projects.\n\nSupported languages:\n  • Rust (Cargo.toml)\n  • Node.js (package.json)\n  • Python (setup.py, pyproject.toml)\n  • Go (go.mod)\n  • Elixir (mix.exs)\n  • C# (.csproj)\n\nFeatures:\n  • Automatic version bumping\n  • Dependency graph resolution\n  • Changelog generation from Git commits\n  • Multi-project coordination\n\nTypical workflow:\n  1. clikd release init\n  2. Make changes and commit\n  3. clikd release status\n  4. clikd release prepare [major|minor|patch]\n  5. Review, commit, tag, and push"
    )]
    Release(ReleaseCommands),

    #[command(
        subcommand,
        about = "Claude AI authentication and changelog generation",
        long_about = "Manage Claude AI authentication for AI-powered changelog generation.\n\nSupports:\n  • Claude Max/Pro subscription via OAuth\n  • API key authentication via ANTHROPIC_API_KEY\n\nUsage:\n  1. clikd ai login - Authenticate with Claude\n  2. clikd release prepare --ai - Generate AI changelog"
    )]
    Ai(AiCommands),
}

#[derive(Args)]
pub struct InitArgs {
    #[arg(long, help = "Generate VSCode settings")]
    pub vscode: bool,

    #[arg(long, help = "Generate IntelliJ/Android Studio settings")]
    pub intellij: bool,

    #[arg(long, help = "Custom working directory")]
    pub workdir: Option<std::path::PathBuf>,
}

#[derive(Subcommand)]
pub enum AuthCommands {
    #[command(about = "Show current authentication status")]
    Status,
}

#[derive(Subcommand)]
pub enum ReleaseCommands {
    #[command(
        about = "Initialize Clikd release management",
        long_about = "Initialize release management in your repository.\n\nThis command:\n  • Detects all projects in your monorepo (Rust, Node.js, Python, Go, Elixir, C#)\n  • Creates .clikd/release.toml configuration\n  • Analyzes project dependencies and builds dependency graph\n  • Sets up changelog tracking\n\nRequires a clean Git working directory unless --force is used."
    )]
    Init {
        #[arg(short, long, help = "Force operation even in unexpected conditions")]
        force: bool,

        #[arg(short, long, help = "The name of the Git upstream remote")]
        upstream: Option<String>,

        #[arg(long, help = "Skip interactive TUI, use automatic detection")]
        no_tui: bool,
    },

    #[command(
        about = "Show release status and changelog",
        long_about = "Display current release status and preview upcoming changes.\n\nShows:\n  • Projects with uncommitted changes\n  • Projects ready for release\n  • Dependency order for releases\n  • Preview of changelog entries based on Git commits\n\nUse this before 'prepare' to verify what will be released."
    )]
    Status {
        #[arg(
            short,
            long,
            value_enum,
            default_value = "table",
            help = "Output format"
        )]
        format: Option<ReleaseOutputFormat>,

        #[arg(long, help = "Force text mode even in TTY")]
        no_tui: bool,
    },

    #[command(
        about = "Prepare a release (bump versions)",
        long_about = "Prepare a new release by bumping versions and updating changelogs.\n\nBump types:\n  • major: Breaking changes (1.0.0 → 2.0.0)\n  • minor: New features (1.0.0 → 1.1.0)\n  • patch: Bug fixes (1.0.0 → 1.0.1)\n  • auto: Automatic bump based on conventional commits\n  • manual: Interactive TUI wizard with suggestions\n\nThis command:\n  • Updates version numbers in all affected project files\n  • Generates/updates CHANGELOG.md for each project\n  • Updates dependency versions in dependent projects\n  • Creates a commit-ready state (you still need to commit and tag)\n\nModes:\n  • TUI mode (default): Interactive 4-step wizard with auto-suggestions\n  • Auto mode (--no-tui): Automatic bump based on conventional commits\n  • Use --no-tui for CI/CD pipelines"
    )]
    Prepare {
        #[arg(help = "Version bump type: major, minor, patch, auto, or manual")]
        bump: Option<String>,

        #[arg(long, help = "Force auto mode, skip interactive TUI wizard")]
        no_tui: bool,

        #[arg(
            short,
            long,
            value_delimiter = ',',
            help = "Per-project version bumps (e.g., gate:major,rig:minor)"
        )]
        project: Option<Vec<String>>,
    },

    #[command(
        about = "Show project dependency graph",
        long_about = "Display the project dependency graph.\n\nInteractive TUI mode (default):\n  • Navigate through projects with arrow keys\n  • View dependency details\n  • Visual dependency tree\n\nBrowser mode (--web):\n  • Interactive Cytoscape.js graph\n  • Multiple layouts (Hierarchy, Force, Circle)\n  • Search, zoom, export PNG\n\nNon-interactive mode (--no-tui):\n  • ASCII art graph\n  • DOT format for Graphviz\n  • JSON for programmatic use"
    )]
    Graph {
        #[arg(short, long, value_enum, help = "Output format (only with --no-tui)")]
        format: Option<GraphOutputFormat>,

        #[arg(long, help = "Skip interactive TUI, output static graph")]
        no_tui: bool,

        #[arg(long, short, help = "Open interactive graph in web browser")]
        web: bool,

        #[arg(long, short, help = "Save HTML graph to file (implies --web)")]
        out: Option<String>,
    },
}

#[derive(Clone, ValueEnum)]
pub enum GraphOutputFormat {
    #[value(help = "ASCII art graph")]
    Ascii,
    #[value(help = "DOT format (for Graphviz)")]
    Dot,
    #[value(help = "JSON format")]
    Json,
}

#[derive(Subcommand)]
pub enum AiCommands {
    #[command(
        about = "Authenticate with Claude Max/Pro subscription",
        long_about = "Authenticate with your Claude Max or Pro subscription.\n\nThis opens a browser window for OAuth authentication.\nAfter logging in, your credentials are stored securely in the system keychain.\n\nAlternatively, set the ANTHROPIC_API_KEY environment variable."
    )]
    Login,

    #[command(about = "Sign out from Claude AI")]
    Logout,

    #[command(about = "Show Claude AI authentication status")]
    Status,
}

#[derive(Args)]
pub struct StartArgs {
    #[arg(
        short = 'x',
        long,
        value_delimiter = ',',
        help = "Exclude specific services from starting (comma-separated)"
    )]
    pub exclude: Option<Vec<String>>,

    #[arg(long, help = "Skip health checks and start immediately")]
    pub ignore_health_check: bool,
}

#[derive(Args)]
pub struct StopArgs {
    #[arg(short, long, help = "Force stop containers immediately")]
    pub force: bool,

    #[arg(long, help = "Remove volumes and clean up all resources")]
    pub purge: bool,
}

#[derive(Args)]
pub struct StatusArgs {
    #[arg(
        short,
        long,
        value_enum,
        default_value = "table",
        help = "Output format (currently unused, TUI mode is always interactive)"
    )]
    pub format: OutputFormat,
}

#[derive(Clone, ValueEnum)]
pub enum OutputFormat {
    #[value(help = "Interactive table (TUI mode)")]
    Table,
    #[value(help = "JSON output")]
    Json,
    #[value(help = "Environment variables")]
    Env,
}

#[derive(Clone, ValueEnum)]
pub enum ReleaseOutputFormat {
    #[value(help = "Interactive table (TUI mode)")]
    Table,
    #[value(help = "Plain text output")]
    Text,
    #[value(help = "JSON output")]
    Json,
}

#[derive(Args)]
pub struct UpdateArgs {
    #[arg(long, help = "Skip confirmation prompts and update immediately")]
    pub yes: bool,
}
