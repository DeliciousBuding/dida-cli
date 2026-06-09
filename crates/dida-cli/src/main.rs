use std::env;
use std::io::{self, Write};

fn main() {
    let args: Vec<String> = env::args().skip(1).collect();
    let result = dida_cli::run(args, option_env!("DIDA_VERSION").unwrap_or("dev"));
    print!("{}", result.stdout);
    let _ = io::stderr().write_all(result.stderr.as_bytes());
    std::process::exit(result.code);
}
