# Tour Editor App

Tour Editor is an HTMX-based tour editing app in Go that implements the following specification.

## Task Overview

Create a web-based editor for tour definition files that allows tour authors to create and edit guided tour experiences. The tours are stored in YAML format with a specific structure that defines locations, narratives, media content, and navigation rules.

## Core Functionality

1. **Tour Basic Information Management**

   - Define tour metadata: ID, name, description, start/end dates, version
   - Set author information and pricing
   - Configure general settings (media caching, preferred language)
   - Upload hero image for the tour
   - Optional milestone (25%, 50%, 75%) completion messages
   - Define farewell message for tour completion

2. **Location Node Management**

   - Create and edit tour nodes (points of interest)
   - Set geographical coordinates (latitude/longitude) if required
   - Add narrative content:
     - Short description
     - Main narrative text
     - Optionally - audio narrative (automatically converted to .ogg format)
   - Attach media files (support for image, audio, video) with:
     - Configurable send delay
     - Optional narrative text for each media file
   - Media file upload or URL reference options

3. **Entry/Exit Conditions**

   Each node can have entry and exit conditions:
   - Types: quiz, Q&A or puzzle:
     - A Q&A will only have a question and an answer.
     - A quiz will have a question with multiple-choice answers (only one is valid).
     - A puzzle has a question and a media file (image, audio, video).
   - Optional hints for each condition type
   - Correct answer must be specified
   - Strict/non-strict mode toggle

4. **Edge (Connection) Management**

   - Connect nodes with directional edges
   - Add transition instructions (text narrative)
   - Configure edge properties:
     - Media files with send delays
     - Silent transition option (flag that disables any media, narrative, conditions options for the Edge)
     - Navigation instructions

5. **Media Handling**

   - Support for various media types (images, video, audio)
   - Automatic conversion only of audio narrative to .ogg format
   - Other media files are uploaded as-is
   - Cloud storage integration (S3) for media files
   - Media file caching options

6. **Data Import/Export**

   - Export tour definitions to YAML format
   - Validate tour structure completeness
   - Ensure all required fields are present

## File Storage and Processing

1. **Media Storage**

   - All media content is uploaded to S3 storage
   - S3 configuration details are stored in the application settings
   - If a URL, rather than a file upload is provided, then the file is downloaded and stored in S3
   - URLs are stored in tour definition
   - Support for both direct file uploads and URL references

2. **Audio Processing**

   - Audio narratives are converted to .ogg format if a presented in a different format
   - Original files can be in any standard audio format
   - Conversion happens before S3 upload
   - Files are converted using ffmpeg library/utility

## Tour Definition Structure

Tours are defined in YAML with a hierarchical structure:
```yaml
tour:
  # Basic Information
  id: string
  name: string
  description: string
  start_date: string
  end_date: string
  version: string
  hero_image: string
  author:
    name: string
    profile_link: string
  price: number # denoted in cents, i.e. 495 for 4.95

  # Nodes (Points of Interest)
  nodes:
    - id: number
      location:
        lat: number
        lon: number
      short_description: string
      narrative: string
      audio_narrative: string
      media_files:
        - type: string
          uri: string
          send_delay: number
          narrative: string
      entry_condition: # optional
        type: string
        strict: boolean
        question: string
        correct_answer: string
        hints: [string]
        options: [string]  # for quiz only
        media_link: string # for puzzle only
      exit_condition: # optional
        type: string
        strict: boolean
        question: string
        correct_answer: string
        hints: [string]
        options: [string]  # for quiz only
        media_link: string # for puzzle only

  # Edges (Connections)
  edges:
    - from: node_id
      to: node_id
      media_files: [same as node media_files]
      condition: [same as node conditions] # optional
      instructions: string
      silent: boolean
```

## Business Rules

### 1. Tour Structure

   - Nodes must have unique IDs
   - Edges must connect existing nodes
   - All required fields must be filled before export

### 2. Node Conditions

   - Entry/exit conditions can be enforced strictly or loosely
   - Questions can have multiple hints
   - A quiz question must have at least two options
   - A puzzle question must have a media file with it

### 3. Media Files

   - Each media file must have a unique identifier
   - Send delays are specified in seconds
   - Media files can have optional narrative text

## Implementation Details

### Project tree

tour-guide-editor/
├── cmd/
│   └── server/
│       └── main.go
├── doc/
├── internal/
│   ├── config/
│   ├── handlers/
│   ├── middleware/
│   ├── mocks/
│   ├── models/
│   ├── services/
│   ├── types/
│   └── validators/
├── templates/
│   ├── editor/
│   └── tour/
├── static/
│   ├── css/
│   └── js/
├── tests/
└── config/

### Requirements

 - [libvips](https://github.com/libvips/libvips)
