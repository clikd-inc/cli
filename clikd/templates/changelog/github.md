# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

{{#each releases}}
## {{title}} {{#if date}}({{date}}){{/if}}
{{#if summary}}
{{summary}}
{{/if}}

{{#each groups}}
### {{title}}

{{#each commits}}
- {{#if id}}{{id}} {{/if}}{{message}}
{{/each}}
{{/each}}
{{/each}}
