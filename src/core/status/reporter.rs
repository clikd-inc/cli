use crate::error::Result;
use crate::cli::OutputFormat;

pub async fn report(_branch: &str, _format: OutputFormat) -> Result<()> {
    Ok(())
}
