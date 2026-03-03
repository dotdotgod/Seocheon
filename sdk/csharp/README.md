# Seocheon .NET SDK (C#)

## Runtime Requirements

- Supported target framework: `net10.0`
- Minimum required runtime: `.NET 10`

Projects consuming this SDK must target `net10.0` (or a compatible newer .NET runtime).

## Build and Test

```bash
dotnet restore
dotnet build -c Release
dotnet test --nologo -v minimal
```
