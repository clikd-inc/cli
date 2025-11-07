use crate::cli::OutputFormat;
use crate::error::Result;

pub async fn report(_branch: &str, _format: OutputFormat) -> Result<()> {
    Ok(())
}
