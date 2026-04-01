-- Seed Professional Resources for MentorMatch Marketplace

-- Ensure the resources table exists with the correct schema
-- (Assuming id, title, category, description, author_name, downloads, rating, file_url, created_at, updated_at)

INSERT INTO public.resources (id, title, category, description, author_name, downloads, rating, file_url, created_at, updated_at)
VALUES 
    (gen_random_uuid(), 'Full-Stack Developer Roadmap 2024', 'roadmaps', 'A comprehensive step-by-step guide to mastering modern web development from HTML to Distributed Systems.', 'Ashish Parshuram', 1240, 4.9, 'https://roadmap.sh/full-stack', NOW(), NOW()),
    
    (gen_random_uuid(), 'FAANG Standard Resume Template', 'templates', 'The exact LaTeX and Google Docs templates used by successful applicants at Google, Meta, and Netflix.', 'Priya Sharma', 850, 4.8, 'https://www.overleaf.com/latex/templates/software-engineer-resume/vytqfymzqxrq', NOW(), NOW()),
    
    (gen_random_uuid(), 'System Design Interview Handbook', 'guides', 'Deep dive into load balancing, caching, databases, and microservices architecture for senior engineering roles.', 'Suresh Kumar', 2100, 5.0, 'https://github.com/donnemartin/system-design-primer', NOW(), NOW()),
    
    (gen_random_uuid(), 'JEE Physics: Rotational Mechanics Hyper-Sheet', 'notes', 'One-page formula sheet covering every concept in Rotational Dynamics for quick revision.', 'Dr. R.K. Yadav', 3420, 4.7, 'https://res.cloudinary.com/demo/image/upload/sample.pdf', NOW(), NOW()),
    
    (gen_random_uuid(), 'Salary Negotiation Script: Tech Edition', 'guides', 'Word-for-word scripts to handle low-ball offers and negotiate an additional $20k-$50k in total compensation.', 'Aman Gupta', 560, 4.9, 'https://res.cloudinary.com/demo/image/upload/sample.pdf', NOW(), NOW()),
    
    (gen_random_uuid(), 'Mock System Design: Uber-like App', 'papers', 'A detailed solution to the "Design Uber" interview question, including sequence diagrams and DB schemas.', 'Vikram Singh', 890, 4.6, 'https://res.cloudinary.com/demo/image/upload/sample.pdf', NOW(), NOW()),
    
    (gen_random_uuid(), 'Machine Learning Interview Questions', 'guides', 'Top 50 most frequently asked questions in ML engineering interviews, from bias-variance to transformer math.', 'Neha Kapoor', 1150, 4.5, 'https://res.cloudinary.com/demo/image/upload/sample.pdf', NOW(), NOW()),
    
    (gen_random_uuid(), 'Product Management Case Study Toolkit', 'templates', 'Frameworks for product-sense, product-execution, and metrics questions used in PM interviews.', 'Anjali Desai', 720, 4.7, 'https://res.cloudinary.com/demo/image/upload/sample.pdf', NOW(), NOW());

-- Update audit timestamps
UPDATE public.resources SET updated_at = NOW() WHERE updated_at IS NULL;
