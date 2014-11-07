require 'digest'
require 'aws-sdk'

BUCKET_NAME = 'heroku-cli'

TARGETS = [
  {os: 'darwin', arch: 'amd64'},
  {os: 'linux', arch: 'amd64'},
  {os: 'linux', arch: '386'},
  {os: 'windows', arch: 'amd64'},
  {os: 'windows', arch: '386'}
]

VERSION = `./version.sh`.chomp
dirty = `git status 2> /dev/null | tail -n1`.chomp != 'nothing to commit, working directory clean'
BRANCH = dirty ? 'dirty' : `git rev-parse --abbrev-ref HEAD`.chomp

puts "hk: #{VERSION}"

task :build do
  FileUtils.mkdir_p 'dist'
  TARGETS.each do |target|
    path = "./dist/hk_#{target[:os]}_#{target[:arch]}"
    puts "building #{path}..."
    build(target[:os], target[:arch], path)
    gzip(path)
  end
end

task :deploy => :build do
  case BRANCH
  when 'dirty'
    deploy('gonpm')
  when 'dev'
    deploy('dev')
  when 'release'
    deploy('release')
  else
    puts 'not on deployable branch (dev/release) current branch is: ' + BRANCH
  end
end

def deploy(channel)
  puts "deploying #{VERSION} to #{channel}..."
  bucket = get_s3_bucket
  TARGETS.each do |target|
    filename = "hk_#{target[:os]}_#{target[:arch]}"
    local_path = "./dist/#{filename}"
    remote_path = "hk/#{channel}/#{VERSION}/#{filename}"
    remote_url = "#{BUCKET_NAME}.s3.amazonaws.com/#{remote_path}"
    puts "uploading #{local_path} to #{remote_url}"
    upload_file(bucket, local_path, remote_path, content_type: 'application/octet-stream')
    upload_file(bucket, local_path + '.gz', remote_path + '.gz', content_type: 'application/gzip', content_encoding: 'gzip')
    upload_string(bucket, sha_digest(local_path), remote_path + ".sha1", content_type: 'text/plain')
  end
  version_path = "hk/#{channel}/VERSION"
  puts "setting #{version_path} to #{VERSION}"
  upload_string(bucket, VERSION, version_path, content_type: 'text/plain')
end

def build(os, arch, path)
  ldflags = "-X main.Version #{VERSION}"
  args = "-o #{path} -ldflags \"#{ldflags}\""
  system("GOOS=#{os} GOARCH=#{arch} go build #{args}")
end

def gzip(path)
  system("gzip --keep -f #{path}")
end

def sha_digest(path)
  Digest::SHA1.file(path).hexdigest
end

def get_s3_bucket
  s3 = AWS::S3.new(region: 'us-west-2', access_key_id: ENV['HEROKU_RELEASE_ACCESS'], secret_access_key: ENV['HEROKU_RELEASE_SECRET'])
  s3.buckets[BUCKET_NAME]
end

def upload_file(bucket, local, remote, opts={})
  obj = bucket.objects[remote]
  obj.write(Pathname.new(local), opts)
  obj.acl = :public_read
end

def upload_string(bucket, s, remote, opts={})
  obj = bucket.objects[remote]
  obj.write(s, opts)
  obj.acl = :public_read
end
