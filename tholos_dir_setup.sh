#!/bin/bash

#Tell user to be in their correct repo
echo "Before running this script please create the repository for the project and clone locally then cd into the top level directory
of the repository and copy this script there locally before running."
echo ""

#Create variables for tholos directory structure and base files setup

#Create Project Name Var
while true; do
    read -p 'What is the project name: ' project_name
    if [[ -z "$project_name" ]]; then
    echo "No input entered for project name, project name is required"
    else break
    fi
done
echo ""

#Create account name Var
while true; do
    read -p 'What is the account designation i.e. dev,prod,uat etc: ' account_name
    if [[ -z "$account_name" ]]; then
    echo "No input entered for account name, account name is required"
    else break
    fi
done
echo ""

#Create account ID Var
while true; do
    read -p 'What is the account ID i.e. 6289628nn: ' account_id
    if [[ -z "$account_id" ]]; then
    echo "No input entered for account ID, account ID is required"
    else break
    fi
done
echo ""

#Create roam-role Var
while true; do
    read -p 'What is the roam role for the account: ' roam_role
    if [[ -z "$roam_role" ]]; then
    echo "No input entered for roam role, roam role is required"
    else break
    fi
done
echo ""

#Create profile Var
while true; do
    read -p 'What is the profile name for the management account as set in your ~/.aws/config ie js-mgmt: ' profile_name
    if [[ -z "$profile_name" ]]; then
    echo "No input entered for profile name, profile name is required"
    else break
    fi
done
echo ""

#Create region Var
while true; do
    read -p 'What region will you run in: ' region_name
    if [[ -z "$region_name" ]]; then
    echo "No input entered for region, region is required"
    else break
    fi
done
echo ""

#Create top level directory
mkdir -p terraform

#Create project file and directories at that level
mkdir -p terraform/modules
mkdir -p terraform/$project_name
echo "project: $project_name" > terraform/project.yaml
echo "region: $region_name" >> terraform/project.yaml
echo "encrypt-s3-state: true" >> terraform/project.yaml
echo "accounts:" >> terraform/project.yaml
echo "  $account_name:" >> terraform/project.yaml
echo "    profile: $profile_name" >> terraform/project.yaml
echo "    roam-role: $roam_role" >> terraform/project.yaml
echo "    account_id: $account_id" >> terraform/project.yaml

#Create development directory
mkdir -p terraform/$project_name/dev
mkdir -p terraform/$project_name/dev/plans
touch terraform/$project_name/dev/plans/.gitkeep

#Create params/plans directory & File
mkdir -p terraform/$project_name/dev/params
touch terraform/$project_name/dev/params/env.tfvars